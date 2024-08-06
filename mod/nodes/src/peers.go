package nodes

import (
	"context"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/nodes/src/frames"
	"github.com/cryptopunkscc/astrald/mod/nodes/src/noise"
	"github.com/cryptopunkscc/astrald/sig"
	"io"
	"slices"
	"sync"
	"sync/atomic"
)

type Peers struct {
	*Module
	streams sig.Set[*Stream]
	conns   sig.Map[astral.Nonce, *conn]
}

func NewPeers(m *Module) *Peers {
	return &Peers{Module: m}
}

func (mod *Peers) RouteQuery(ctx context.Context, query *astral.Query, caller io.WriteCloser) (w io.WriteCloser, err error) {
	if !mod.isRoutable(query.Target) {
		return astral.RouteNotFound(mod)
	}

	conn, ok := mod.conns.Set(query.Nonce, newConn(query.Nonce))
	if !ok {
		return astral.RouteNotFound(mod, errors.New("nonce already exists"))
	}

	conn.RemoteIdentity = query.Target
	conn.Query = query.Query
	conn.Outbound = true

	// make sure we're linked with the target node
	if err := mod.ensureConnected(ctx, query.Target); err != nil {
		conn.swapState(stateRouting, stateClosed)
		return astral.RouteNotFound(mod, err)
	}

	// prepare the protocol frame
	frame := &frames.Query{
		Nonce:  query.Nonce,
		Query:  query.Query,
		Buffer: uint32(conn.rsize),
	}

	// send the query via all streams
	for _, s := range mod.streams.Select(func(s *Stream) bool {
		return s.RemoteIdentity().IsEqual(query.Target)
	}) {
		go s.Write(frame)
	}

	// wait for the response
	select {
	case accepted := <-conn.res:
		if !accepted {
			return astral.Reject()
		}

		go func() {
			io.Copy(caller, conn)
			caller.Close()
		}()

		return conn, nil

	case <-ctx.Done():
		conn.swapState(stateRouting, stateClosed)
		return astral.RouteNotFound(mod, ctx.Err())
	}
}

func (mod *Peers) peers() (peers []*astral.Identity) {
	var r map[string]struct{}

	for _, s := range mod.streams.Clone() {
		if _, found := r[s.RemoteIdentity().String()]; found {
			continue
		}
		r[s.RemoteIdentity().String()] = struct{}{}
		peers = append(peers, s.RemoteIdentity())
	}

	return
}

func (mod *Peers) frameReader(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case frame := <-mod.in:
			switch f := frame.Frame.(type) {
			case *frames.Query:
				go mod.handleQuery(frame.Source, f)
			case *frames.Response:
				mod.handleResponse(frame.Source, f)
			case *frames.Ping:
				mod.handlePing(frame.Source, f)
			case *frames.Data:
				mod.handleData(frame.Source, f)
			case *frames.Reset:
				mod.handleReset(frame.Source, f)
			case *frames.Read:
				mod.handleRead(frame.Source, f)
			default:
				mod.log.Errorv(2, "unknown frame: %v", frame.Frame)
			}
		}
	}
}

func (mod *Peers) handleQuery(s *Stream, f *frames.Query) {
	conn, ok := mod.conns.Set(f.Nonce, newConn(f.Nonce))
	if !ok {
		return // ignore duplicates
	}

	conn.RemoteIdentity = s.RemoteIdentity()
	conn.Query = f.Query
	conn.stream = s
	conn.wsize = int(f.Buffer)

	var q = &astral.Query{
		Nonce:  f.Nonce,
		Caller: s.RemoteIdentity(),
		Target: s.LocalIdentity(),
		Query:  f.Query,
	}

	q.Extra.Set("origin", astral.OriginNetwork)

	err := mod.provider.PreprocessQuery(q)
	if err != nil {
		panic(err)
	}

	w, err := mod.provider.RouteQuery(context.Background(), q, conn)
	if err != nil {
		w, err = mod.node.RouteQuery(context.Background(), q, conn)
	}

	if err != nil {
		conn.swapState(stateRouting, stateClosed)
		s.Write(&frames.Response{Nonce: f.Nonce, ErrCode: frames.CodeRejected})
		return
	}

	s.Write(&frames.Response{Nonce: f.Nonce, ErrCode: frames.CodeAccepted, Buffer: uint32(conn.rsize)})
	conn.swapState(stateRouting, stateOpen)

	go func() {
		io.Copy(w, conn)
		w.Close()
	}()
}

func (mod *Peers) handleResponse(s *Stream, f *frames.Response) {
	// find the connection
	conn, ok := mod.conns.Get(f.Nonce)
	if !ok {
		return
	}

	// make sure we sent the query to the identity that sent the response
	if !conn.RemoteIdentity.IsEqual(s.RemoteIdentity()) {
		return
	}

	// if rejected
	if f.ErrCode != 0 {
		if !conn.swapState(stateRouting, stateClosed) {
			return
		}
		conn.res <- false
	}

	if !conn.swapState(stateRouting, stateOpen) {
		return
	}
	conn.stream = s
	conn.wsize = int(f.Buffer)
	conn.res <- true
}

func (mod *Peers) handleData(s *Stream, f *frames.Data) {
	conn, ok := mod.conns.Get(f.Nonce)
	if !ok {
		s.Write(&frames.Reset{Nonce: f.Nonce})
		return
	}

	if conn.state.Load() != stateOpen {
		s.Write(&frames.Reset{Nonce: f.Nonce})
		return
	}

	err := conn.pushRead(f.Payload)
	if err != nil {
		conn.Close()
		return
	}
}

func (mod *Peers) handleRead(s *Stream, f *frames.Read) {
	conn, ok := mod.conns.Get(f.Nonce)
	if !ok {
		s.Write(&frames.Reset{Nonce: f.Nonce})
		return
	}

	conn.growRemoteBuffer(int(f.Len))
}

func (mod *Peers) handleReset(s *Stream, f *frames.Reset) {
	conn, ok := mod.conns.Get(f.Nonce)
	if !ok {
		return
	}

	conn.swapState(stateOpen, stateClosed)
}

func (mod *Peers) handlePing(s *Stream, f *frames.Ping) {
	if f.Pong {
		rtt, err := s.pong(f.Nonce)
		if err != nil {
			mod.log.Errorv(1, "invalid pong nonce from %v", s.RemoteIdentity())
		} else {
			if mod.config.LogPings {
				mod.log.Logv(1, "ping with %v: %v", s.RemoteIdentity(), rtt)
			}
		}
	} else {
		s.Write(&frames.Ping{
			Nonce: f.Nonce,
			Pong:  true,
		})
	}
}

func (mod *Peers) addStream(s *Stream) (err error) {
	linked := mod.isLinked(s.RemoteIdentity())

	err = mod.streams.Add(s)
	if err == nil {
		if !linked {
			mod.Objects.PushLocal(&nodes.EventLinked{NodeID: s.RemoteIdentity()})
		}

		mod.log.Infov(1, "stream with %v added", s.RemoteIdentity())
		go func() {
			for frame := range s.Read() {
				mod.in <- &Frame{
					Frame:  frame,
					Source: s,
				}
			}
			mod.log.Errorv(1, "stream with %v removed: %v", s.RemoteIdentity(), s.Err())
			mod.streams.Remove(s)
			for _, c := range mod.conns.Select(func(k astral.Nonce, v *conn) (ok bool) {
				return v.stream == s
			}) {
				c.Close()
			}

			if !mod.isLinked(s.RemoteIdentity()) {
				mod.Objects.PushLocal(&nodes.EventUnlinked{NodeID: s.RemoteIdentity()})
			}
		}()
	}

	return
}

func (mod *Peers) isLinked(remoteID *astral.Identity) bool {
	for _, s := range mod.streams.Clone() {
		if s.RemoteIdentity().IsEqual(remoteID) {
			return true
		}
	}
	return false
}

func (mod *Peers) isRoutable(identity *astral.Identity) bool {
	return mod.isLinked(identity) || mod.hasEndpoints(identity)
}

func (mod *Peers) Connect(ctx context.Context, remoteID *astral.Identity, conn exonet.Conn) (link io.Closer, err error) {
	defer func() {
		if err != nil {
			conn.Close()
		}
	}()

	aconn, err := noise.HandshakeOutbound(ctx, conn, remoteID, mod.node.Identity())
	if err != nil {
		return nil, fmt.Errorf("outbound handshake: %w", err)
	}

	var linkFeatures []string

	err = cslq.Decode(aconn, "[s][c]c", &linkFeatures)
	if err != nil {
		return nil, fmt.Errorf("read features: %w", err)
	}

	if slices.Contains(linkFeatures, featureMux2) {
		err = cslq.Encode(aconn, "[c]c", featureMux2)
		if err != nil {
			return nil, fmt.Errorf("write: %w", err)
		}

		var errCode int
		err = cslq.Decode(aconn, "c", &errCode)
		if errCode != 0 {
			return nil, errors.New("link feature negotation error")
		}

		mod.addStream(newStream(aconn, true))

		return nil, err
	}

	return nil, errors.New("no supported link types found")
}

func (mod *Peers) Accept(ctx context.Context, conn exonet.Conn) (err error) {
	defer func() {
		if err != nil {
			conn.Close()
		}
	}()

	aconn, err := noise.HandshakeInbound(ctx, conn, mod.node.Identity())
	if err != nil {
		return
	}

	var linkFeatures = []string{featureMux2}

	err = cslq.Encode(aconn, "[s][c]c", linkFeatures)
	if err != nil {
		return
	}

	for {
		var feature string
		err = cslq.Decode(aconn, "[c]c", &feature)
		if err != nil {
			return
		}

		switch feature {
		case featureMux2:
			err = cslq.Encode(aconn, "c", 0)
			if err == nil {
				mod.addStream(newStream(aconn, false))
			}

			return

		default:
			cslq.Encode(aconn, "c", 1)
			return fmt.Errorf("remote party (%s from %s) requested an invalid feature: %s",
				aconn.RemoteIdentity(),
				aconn.RemoteEndpoint(),
				feature,
			)
		}
	}
}

func (mod *Peers) connectAt(ctx context.Context, remoteIdentity *astral.Identity, e exonet.Endpoint) error {
	conn, err := mod.Exonet.Dial(ctx, e)
	if err != nil {
		return err
	}

	_, err = mod.Connect(ctx, remoteIdentity, conn)
	if err != nil {
		return err
	}

	return nil
}

func (mod *Peers) connectAny(ctx context.Context, remoteIdentity *astral.Identity, endpoints []exonet.Endpoint) error {
	var queue = sig.ArrayToChan(endpoints)

	if len(queue) == 0 {
		return errors.New("no endpoints provided")
	}

	var wg sync.WaitGroup
	var success atomic.Bool
	var workers = DefaultWorkerCount

	wctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for {
				var e exonet.Endpoint
				var ok bool

				select {
				case <-wctx.Done():
					return
				case e, ok = <-queue:
					if !ok {
						return
					}
				}

				err := mod.connectAt(wctx, remoteIdentity, e)
				if err == nil {
					success.Store(true)
					cancel()
					return
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		cancel()
	}()

	<-wctx.Done()
	if success.Load() {
		return nil
	}

	return errors.New("no endpoint could be reached")
}

func (mod *Peers) ensureConnected(ctx context.Context, remoteIdentity *astral.Identity) error {
	if mod.isLinked(remoteIdentity) {
		return nil
	}

	return mod.connectAny(ctx, remoteIdentity, mod.Endpoints(remoteIdentity))
}
