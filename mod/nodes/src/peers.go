package nodes

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/nodes/src/frames"
	"github.com/cryptopunkscc/astrald/mod/nodes/src/noise"
	"github.com/cryptopunkscc/astrald/sig"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

type Peers struct {
	*Module
	streams  sig.Set[*Stream]
	sessions sig.Map[astral.Nonce, *session]
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
	conn, ok := mod.sessions.Set(q.Nonce, newSession(q.Nonce))
	if !ok {
		return query.RouteNotFound(mod, errors.New("sessionId already exists"))
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
			mod.sessions.Delete(q.Nonce)
			return query.RejectWithCode(errCode)
		}

		go func() {
			io.Copy(w, conn)
			w.Close()
		}()

		return conn, nil

	case <-ctx.Done():
		conn.swapState(stateRouting, stateClosed)
		mod.sessions.Delete(q.Nonce)
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
			case *frames.Migrate:
				mod.handleMigrate(frame.Source, f)
			default:
				mod.log.Errorv(2, "unknown frame: %v", frame.Frame)
			}
		}
	}
}

func (mod *Peers) handleQuery(s *Stream, f *frames.Query) {
	conn, ok := mod.sessions.Set(f.Nonce, newSession(f.Nonce))
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
	conn, ok := mod.sessions.Get(f.Nonce)
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
	session, ok := mod.sessions.Get(f.Nonce)
	if !ok {
		s.Write(&frames.Reset{Nonce: f.Nonce})
		return
	}

	switch session.state.Load() {
	case stateOpen, stateMigrating:
	default:
		mod.log.Errorv(1, "received data frame from %v in state %v", s.RemoteIdentity(), session.state.Load())
		s.Write(&frames.Reset{Nonce: f.Nonce})
		return
	}

	err := session.pushRead(f.Payload)
	if err != nil {
		mod.log.Errorv(1, "failed to push read frame: %v", err)
		session.Close()
		return
	}
}

func (mod *Peers) handleRead(s *Stream, f *frames.Read) {
	conn, ok := mod.sessions.Get(f.Nonce)
	if !ok {
		s.Write(&frames.Reset{Nonce: f.Nonce})
		return
	}

	conn.growRemoteBuffer(int(f.Len))
}

func (mod *Peers) handleReset(s *Stream, f *frames.Reset) {
	conn, ok := mod.sessions.Get(f.Nonce)
	if !ok {
		return
	}

	conn.swapState(stateOpen, stateClosed)
}

func (mod *Peers) handlePing(s *Stream, f *frames.Ping) {
	if f.Pong {
		rtt, err := s.pong(f.Nonce)
		if err != nil {
			mod.log.Errorv(1, "invalid pong sessionId from %v", s.RemoteIdentity())
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

func (mod *Peers) handleMigrate(s *Stream, f *frames.Migrate) {
	conn, ok := mod.sessions.Get(f.Nonce)
	if !ok {
		return
	}

	_ = conn.CompleteMigration()
}

func (mod *Peers) addStream(
	s *Stream,
) (err error) {
	var (
		dir     = "in"
		netName = "unknown network"
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
	streamsWithSameIdentity := mod.streams.Select(func(v *Stream) bool {
		return v.RemoteIdentity().IsEqual(s.RemoteIdentity())
	})

	if !s.outbound {
		mod.linkPool.notifyStreamWatchers(s, nil)
	}

	mod.Events.Emit(&nodes.StreamCreatedEvent{
		RemoteIdentity: s.RemoteIdentity(),
		StreamId:       s.id,
		StreamCount:    len(streamsWithSameIdentity),
	})

	// handle the stream
	go func() {
		mod.readStreamFrames(s)

		// remove the stream and its connections
		mod.streams.Remove(s)
		for _, c := range mod.sessions.Select(func(k astral.Nonce, v *session) (ok bool) {
			return v.stream == s
		}) {
			c.Close()
		}

		streamsWithSameIdentity := mod.streams.Select(func(v *Stream) bool {
			return v.RemoteIdentity().IsEqual(s.RemoteIdentity())
		})

		mod.Events.Emit(&nodes.StreamClosedEvent{
			RemoteIdentity: s.RemoteIdentity(),
			Forced:         false,
			StreamCount:    astral.Int8(len(streamsWithSameIdentity)),
		})

		mod.log.Info("closed %v-stream with %v (%v): %v", dir, s.RemoteIdentity(), netName, s.Err())
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

// negotiateOutboundStream reads peer's supported features and the session sessionId.
func (mod *Peers) negotiateOutboundStream(aconn astral.Conn) (features []astral.String8, err error) {
	var featCount astral.Uint16
	if _, err = featCount.ReadFrom(aconn); err != nil {
		return nil, fmt.Errorf("read features: %w", err)
	}
	for i := 0; i < int(featCount); i++ {
		var feat astral.String8
		if _, err = feat.ReadFrom(aconn); err != nil {
			return nil, fmt.Errorf("read features: %w", err)
		}
		features = append(features, feat)
	}

	return features, nil
}

// negotiateInboundStream sends our supported features and a fresh session sessionId.
func (mod *Peers) negotiateInboundStream(aconn astral.Conn) (err error) {
	var linkFeatures = []string{featureMux2}
	if _, err = astral.Uint16(len(linkFeatures)).WriteTo(aconn); err != nil {
		return err
	}
	for _, feature := range linkFeatures {
		if _, err = astral.String8(feature).WriteTo(aconn); err != nil {
			return err
		}
	}

	return nil
}

func (mod *Peers) setInboundStreamNonce(aconn astral.Conn) (nonce astral.Nonce, err error) {
	nonce = astral.NewNonce()
	if _, err = nonce.WriteTo(aconn); err != nil {
		return 0, err
	}

	return nonce, nil
}

func (mod *Peers) readOutboundStreamNonce(aconn astral.Conn) (nonce astral.Nonce, err error) {
	if _, err = nonce.ReadFrom(aconn); err != nil {
		return 0, err
	}

	return nonce, nil
}

func (mod *Peers) EstablishOutboundLink(ctx context.Context, remoteID *astral.Identity, conn exonet.Conn) (_ *Stream, err error) {
	defer func() {
		if err != nil {
			conn.Close()
		}
	}()

	// get the node's private key
	privKey, err := mod.getPrivateKey()
	if err != nil {
		return nil, err
	}

	// initiate a handshake
	aconn, err := noise.HandshakeOutbound(
		ctx,
		conn,
		remoteID.PublicKey(),
		secp256k1.PrivKeyFromBytes(privKey.Key),
	)
	if err != nil {
		return nil, fmt.Errorf("outbound handshake: %w", err)
	}

	// Read peer features and session sessionId
	linkFeatures, err := mod.negotiateOutboundStream(aconn)
	if err != nil {
		return nil, err
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

		nonce, err := mod.readOutboundStreamNonce(aconn)
		if err != nil {
			return nil, fmt.Errorf(`read outbound stream nonce: %w`, err)
		}

		stream := newStream(aconn, nonce, true)
		err = mod.addStream(stream)

		return stream, err
	}

	return nil, errors.New("no supported link types found")
}

func (mod *Peers) EstablishInboundLink(ctx context.Context, conn exonet.Conn) (err error) {
	defer func() {
		if err != nil {
			conn.Close()
		}
	}()

	// get the node's private key
	privKey, err := mod.getPrivateKey()
	if err != nil {
		return err
	}

	// respond to a handshake
	aconn, err := noise.HandshakeInbound(ctx, conn, secp256k1.PrivKeyFromBytes(privKey.Key))
	if err != nil {
		return
	}

	// Send our features and session sessionId
	err = mod.negotiateInboundStream(aconn)
	if err != nil {
		return err
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
				nonce, err := mod.setInboundStreamNonce(aconn)
				if err != nil {
					return fmt.Errorf("failed to set inbound stream nonce: %w", err)
				}

				stream := newStream(aconn, nonce, false)
				err = mod.addStream(stream)
				if err != nil {
					return fmt.Errorf("failed to add stream: %w", err)
				}
			}

			return err
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
