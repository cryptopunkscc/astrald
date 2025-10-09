package nodes

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/nodes/src/frames"
	"github.com/cryptopunkscc/astrald/mod/nodes/src/noise"
	"github.com/cryptopunkscc/astrald/sig"
)

type Peers struct {
	*Module
	streams sig.Set[*Stream]
	conns   sig.Map[astral.Nonce, *conn]
}

func NewPeers(m *Module) *Peers {
	return &Peers{Module: m}
}

func (mod *Peers) RouteQuery(ctx *astral.Context, q *astral.Query, w io.WriteCloser) (_ io.WriteCloser, err error) {
	streams := mod.streams.Select(func(s *Stream) bool {
		return s.RemoteIdentity().IsEqual(q.Target)
	})

	// are we linked?
	if len(streams) == 0 {
		return query.RouteNotFound(mod)
	}

	// prepare the connection info
	conn, ok := mod.conns.Set(q.Nonce, newConn(q.Nonce))
	if !ok {
		return query.RouteNotFound(mod, errors.New("nonce already exists"))
	}

	conn.RemoteIdentity = q.Target
	conn.Query = q.Query
	conn.Outbound = true

	// prepare the query frame
	frame := &frames.Query{
		Nonce:  q.Nonce,
		Query:  q.Query,
		Buffer: uint32(conn.rsize),
	}

	// send the query via all streams
	for _, s := range streams {
		go s.Write(frame)
	}

	// wait for the response
	select {
	case errCode := <-conn.res:
		if errCode != 0 {
			mod.conns.Delete(q.Nonce)
			return query.RejectWithCode(errCode)
		}

		go func() {
			io.Copy(w, conn)
			w.Close()
		}()

		return conn, nil

	case <-ctx.Done():
		conn.swapState(stateRouting, stateClosed)
		mod.conns.Delete(q.Nonce)
		return query.RouteNotFound(mod, ctx.Err())
	}
}

func (mod *Peers) peers() (peers []*astral.Identity) {
	var r = map[string]struct{}{}

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

	ctx := astral.NewContext(nil).WithIdentity(mod.node.Identity())

	w, err := mod.node.RouteQuery(ctx, q, conn)
	if err != nil {
		conn.swapState(stateRouting, stateClosed)
		var code = uint8(frames.CodeRejected)
		var reject *astral.ErrRejected
		if errors.As(err, &reject) {
			code = reject.Code
		}
		s.Write(&frames.Response{Nonce: f.Nonce, ErrCode: code})
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
		conn.res <- f.ErrCode
	}

	if !conn.swapState(stateRouting, stateOpen) {
		return
	}
	conn.stream = s
	conn.wsize = int(f.Buffer)
	conn.res <- 0
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

func (mod *Peers) addStream(
	s *Stream,
) (err error) {
	var (
		alreadyLinked = mod.isLinked(s.RemoteIdentity())
		dir           = "in"
		netName       = "unknown network"
	)

	if s.outbound {
		dir = "out"
	}

	// try to figure out the network name
	switch {
	case s.LocalEndpoint() != nil:
		netName = s.LocalEndpoint().Network()
	case s.RemoteEndpoint() != nil:
		netName = s.RemoteEndpoint().Network()
	}

	err = mod.streams.Add(s)
	if err != nil {
		return
	}

	// log stream addition
	mod.log.Infov(1, "added %v-stream with %v (%v)", dir, s.RemoteIdentity(), netName)

	// emit an event if linked
	if !alreadyLinked {
		mod.Objects.Receive(&nodes.EventLinked{NodeID: s.RemoteIdentity()}, nil)
	}

	// handle the stream
	go func() {
		mod.readStreamFrames(s)

		// remove the stream and its connections
		mod.streams.Remove(s)
		for _, c := range mod.conns.Select(func(k astral.Nonce, v *conn) (ok bool) {
			return v.stream == s
		}) {
			c.Close()
		}

		// log stream removal
		mod.log.Errorv(1, "removed %v-stream with %v (%v): %v", dir, s.RemoteIdentity(), netName, s.Err())

		// emit an event if unlinked
		if !mod.isLinked(s.RemoteIdentity()) {
			mod.Objects.Receive(&nodes.EventUnlinked{NodeID: s.RemoteIdentity()}, nil)
		}
	}()

	// reflect the stream
	go mod.reflectStream(s)

	return
}

// reflectStream reflects the observed remote endpoint on inbound streams
func (mod *Peers) reflectStream(s *Stream) (err error) {
	// reflect only on inbound streams with a known remote endpoint
	if s.outbound || s.RemoteEndpoint() == nil {
		return
	}

	// reflect the endpoint
	err = mod.Objects.Push(mod.ctx, s.RemoteIdentity(),
		&nodes.ObservedEndpointMessage{
			Endpoint: s.RemoteEndpoint(),
		})

	// log the result
	if err != nil {
		mod.log.Errorv(2, "Objects.Push(%v, %v): %v", s.RemoteIdentity(), s.RemoteEndpoint(), err)
	} else {
		mod.log.Logv(2, "reflected endpoint %v to %v", s.RemoteEndpoint(), s.RemoteIdentity())
	}

	return
}

// readStreamFrames reads frames from the stream until it closes
func (mod *Peers) readStreamFrames(s *Stream) {
	// read frames
	for frame := range s.Read() {
		mod.in <- &Frame{
			Frame:  frame, // NOTE: add timeout handling?
			Source: s,
		}
	}
}

// isLinked returns true if there's at least one stream with the given remote identity
func (mod *Peers) isLinked(remoteID *astral.Identity) bool {
	for _, s := range mod.streams.Clone() {
		if s.RemoteIdentity().IsEqual(remoteID) {
			return true
		}
	}
	return false
}

func (mod *Peers) Connect(ctx context.Context, remoteID *astral.Identity, conn exonet.Conn) (_ *Stream, err error) {
	defer func() {
		if err != nil {
			conn.Close()
		}
	}()

	aconn, err := noise.HandshakeOutbound(ctx, conn, remoteID, mod.node.Identity())
	if err != nil {
		return nil, fmt.Errorf("outbound handshake: %w", err)
	}

	var linkFeatures []astral.String
	var featCount astral.Uint16

	_, err = featCount.ReadFrom(aconn)
	if err != nil {
		return nil, fmt.Errorf("read features: %w", err)
	}

	for i := 0; i < int(featCount); i++ {
		var feat astral.String8
		_, err = feat.ReadFrom(aconn)
		if err != nil {
			return nil, fmt.Errorf("read features: %w", err)
		}
		linkFeatures = append(linkFeatures, astral.String(feat))
	}

	if slices.Contains(linkFeatures, featureMux2) {
		_, err = astral.String8(featureMux2).WriteTo(aconn)
		if err != nil {
			return nil, fmt.Errorf("write: %w", err)
		}

		var errCode astral.Int8
		_, err = errCode.ReadFrom(aconn)
		switch {
		case err != nil:
			return nil, fmt.Errorf("read: %w", err)
		case errCode != 0:
			return nil, errors.New("link feature negotation error")
		}

		stream := newStream(aconn, true)

		err = mod.addStream(stream)

		return stream, err
	}

	return nil, errors.New("no supported link types found")
}

func (mod *Peers) Accept(ctx *astral.Context, conn exonet.Conn) (err error) {
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

	_, err = astral.Uint16(len(linkFeatures)).WriteTo(aconn)
	if err != nil {
		return
	}

	for _, feature := range linkFeatures {
		_, err = astral.String8(feature).WriteTo(aconn)
		if err != nil {
			return
		}
	}

	for {
		var feature string
		_, err = (*astral.String8)(&feature).ReadFrom(aconn)
		if err != nil {
			return
		}

		switch feature {
		case featureMux2:
			_, err = astral.Uint8(0).WriteTo(aconn)
			if err == nil {
				mod.addStream(newStream(aconn, false))
				err = mod.pushObservedEndpoint(ctx, aconn)
				if err != nil {
					return err
				}
			}

			return

		default:
			_, err = astral.Uint8(1).WriteTo(aconn)
			return fmt.Errorf("remote party (%s from %s) requested an invalid feature: %s",
				aconn.RemoteIdentity(),
				aconn.RemoteEndpoint(),
				feature,
			)
		}
	}
}

func (mod *Peers) pushObservedEndpoint(
	ctx *astral.Context,
	conn *noise.Conn,
) (err error) {
	var endpoint = conn.RemoteEndpoint()

	var remoteIdentity = conn.RemoteIdentity()

	err = mod.Objects.Push(ctx, remoteIdentity, &nodes.ObservedEndpointEvent{
		Endpoint: endpoint,
	})
	if err != nil {
		return fmt.Errorf("nodes peers/push failed: %w", err)
	}

	return nil
}

func (mod *Peers) connectAt(ctx *astral.Context, remoteIdentity *astral.Identity, e exonet.Endpoint) (*Stream, error) {
	conn, err := mod.Exonet.Dial(ctx, e)
	if err != nil {
		return nil, err
	}

	return mod.Connect(ctx, remoteIdentity, conn)
}

func (mod *Peers) connectAtAny(ctx *astral.Context, remoteIdentity *astral.Identity, endpoints <-chan exonet.Endpoint) (*Stream, error) {
	var wg sync.WaitGroup
	var out sig.Value[*Stream]
	var workers = DefaultWorkerCount

	wctx, cancel := ctx.WithCancel()
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
				case e, ok = <-endpoints:
					if !ok {
						return
					}
				}

				stream, err := mod.connectAt(wctx, remoteIdentity, e)
				if err != nil {
					continue
				}

				if _, ok := out.Swap(nil, stream); ok {
					cancel()
				} else {
					stream.CloseWithError(errors.New("excess stream"))
				}

				return
			}
		}()
	}

	go func() {
		wg.Wait()
		cancel()
	}()

	<-wctx.Done()

	stream := out.Get()
	if stream != nil {
		return stream, nil
	}

	return nil, errors.New("no endpoint could be reached")
}
