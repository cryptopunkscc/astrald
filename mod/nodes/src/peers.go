package nodes

import (
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/gateway"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/nodes/frames"
	"github.com/cryptopunkscc/astrald/mod/nodes/src/noise"
	"github.com/cryptopunkscc/astrald/sig"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

type Peers struct {
	*Module
	links    sig.Set[*Link]
	sessions sig.Map[astral.Nonce, *session]
}

func NewPeers(m *Module) *Peers {
	return &Peers{Module: m}
}

func (mod *Peers) RouteQuery(ctx *astral.Context, q *astral.InFlightQuery, w io.WriteCloser) (_ io.WriteCloser, err error) {
	links := mod.links.Select(func(link *Link) bool {
		return link.RemoteIdentity().IsEqual(q.Target)
	})

	// are we linked?
	if len(links) == 0 {
		return query.RouteNotFound()
	}

	// prepare the connection info
	conn, ok := mod.createSession(q.Nonce)
	if !ok {
		return query.RouteNotFound()
	}

	conn.RemoteIdentity = q.Target
	conn.Query = q.QueryString
	conn.Outbound = true

	// prepare the query frame
	frame := &frames.Query{
		Nonce:  q.Nonce,
		Query:  q.QueryString,
		Buffer: uint32(defaultBufferSize),
	}

	// send the query via all links
	for _, link := range links {
		go link.Write(frame)
	}

	// wait for the response
	select {
	case errCode := <-conn.res:
		if errCode != 0 {
			return query.RejectWithCode(errCode)
		}

		go func() {
			io.Copy(w, conn)
			w.Close()
		}()

		return conn, nil

	case <-ctx.Done():
		conn.Close()
		return query.RouteNotFound()
	}
}

func (mod *Peers) peers() (peers []*astral.Identity) {
	var r = map[string]struct{}{}

	for _, link := range mod.links.Clone() {
		if _, found := r[link.RemoteIdentity().String()]; found {
			continue
		}
		r[link.RemoteIdentity().String()] = struct{}{}
		peers = append(peers, link.RemoteIdentity())
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

func (mod *Peers) handleQuery(link *Link, f *frames.Query) {
	mod.handleInboundQuery(link, f.Nonce, link.RemoteIdentity(), link.LocalIdentity(), nil, f.Query, int(f.Buffer))
}

// handleRelayQuery routes a query received via a relay channel to its target and bridges
// the resulting session back to the caller over the relay peer's stream.
func (mod *Peers) handleRelayQuery(link *Link, relayQuery *nodes.QueryContainer) error {
	mod.handleInboundQuery(
		link,
		relayQuery.Query.Nonce,
		relayQuery.CallerID,
		relayQuery.TargetID,
		link.RemoteIdentity(),
		relayQuery.Query.Query,
		int(relayQuery.Query.Buffer),
	)

	return nil
}

func (mod *Peers) handleInboundQuery(link *Link, nonce astral.Nonce, caller, target *astral.Identity, relayID *astral.Identity, queryStr string, initBuffer int) {
	conn, ok := mod.createSession(nonce)
	if !ok {
		return // ignore duplicates
	}

	conn.RemoteIdentity = caller
	conn.relayID = relayID
	conn.Query = queryStr
	conn.link = link

	var q = astral.Launch(&astral.Query{
		Nonce:       nonce,
		Caller:      caller,
		Target:      target,
		QueryString: queryStr,
	})

	q.Extra.Set("origin", astral.OriginNetwork)
	ctx := astral.NewContext(nil).WithIdentity(mod.node.Identity())

	w, err := mod.node.RouteQuery(ctx, q, conn)
	if err != nil {
		conn.Close()
		var code = uint8(frames.CodeRejected)
		var reject *astral.ErrRejected
		if errors.As(err, &reject) {
			code = reject.Code
		}
		link.Write(&frames.Response{Nonce: nonce, ErrCode: code})
		return
	}

	resetFunc := func() { link.Write(&frames.Reset{Nonce: nonce}) }
	reader := newSessionReader(mod.newMuxInputBuffer(link, nonce))
	writer := newSessionWriter(mod.newMuxOutputBuffer(link, nonce, conn), resetFunc)
	writer.Grow(initBuffer)

	if err := conn.Setup(link, reader, writer); err != nil {
		return
	}
	link.Write(&frames.Response{Nonce: nonce, ErrCode: frames.CodeAccepted, Buffer: uint32(defaultBufferSize)})
	conn.Open()

	go func() {
		io.Copy(w, conn)
		w.Close()
	}()
}

func (mod *Peers) handleResponse(link *Link, f *frames.Response) {
	conn, ok := mod.sessions.Get(f.Nonce)
	if !ok {
		return
	}

	expectedSource := conn.RemoteIdentity
	if conn.relayID != nil {
		expectedSource = conn.relayID
	}

	if !expectedSource.IsEqual(link.RemoteIdentity()) {
		return
	}

	// if rejected
	if f.ErrCode != 0 {
		if !conn.swapState(stateRouting, stateClosed) {
			return
		}
		conn.res <- f.ErrCode
		return
	}

	if !conn.swapState(stateRouting, stateOpen) {
		return
	}

	resetFunc := func() { link.Write(&frames.Reset{Nonce: f.Nonce}) }
	reader := newSessionReader(mod.newMuxInputBuffer(link, f.Nonce))
	writer := newSessionWriter(mod.newMuxOutputBuffer(link, f.Nonce, conn), resetFunc)
	writer.Grow(int(f.Buffer))
	if err := conn.Setup(link, reader, writer); err != nil {
		return
	}
	conn.Open()

	conn.res <- 0
}

func (mod *Peers) handleData(link *Link, f *frames.Data) {
	session, ok := mod.sessions.Get(f.Nonce)
	if !ok {
		link.Write(&frames.Reset{Nonce: f.Nonce})
		return
	}

	switch session.getState() {
	case stateOpen, stateMigrating:
	default:
		mod.log.Errorv(1, "received data frame from %v in state %v", link.RemoteIdentity(), session.getState())
		link.Write(&frames.Reset{Nonce: f.Nonce})
		return
	}

	link.throughput.Add(uint64(len(f.Payload)))
	if link.pressure != nil {
		link.pressure.OnBytes(len(f.Payload), time.Now())
	}

	reader, ok := session.reader.(*muxSessionReader)
	if !ok {
		mod.log.Errorv(1, "received data frame from %v on non-mux session", link.RemoteIdentity())
		return
	}

	if err := reader.Push(f.Payload); err != nil {
		mod.log.Errorv(1, "failed to push read frame: %v", err)
		session.Close()
		return
	}
}

func (mod *Peers) handleRead(link *Link, f *frames.Read) {
	session, ok := mod.sessions.Get(f.Nonce)
	if !ok {
		link.Write(&frames.Reset{Nonce: f.Nonce})
		return
	}

	// Discard Read frames from the old stream after migration has swapped
	// the session to a new stream. These credits correspond to the old
	// input buffer and must not inflate the new output buffer's wsize.
	if !session.isOnLink(link) {
		return
	}

	writer, ok := session.writer.(*muxSessionWriter)
	if ok {
		writer.Grow(int(f.Len))
	}
}

func (mod *Peers) handleReset(link *Link, f *frames.Reset) {
	session, ok := mod.sessions.Get(f.Nonce)
	if !ok {
		return
	}

	if w, ok := session.writer.(*muxSessionWriter); ok {
		w.PeerClose()
	}
	session.Close()
}

func (mod *Peers) handlePing(link *Link, f *frames.Ping) {
	if f.Pong {
		rtt, err := link.pong(f.Nonce)
		if err != nil {
			mod.log.Errorv(1, "invalid pong sessionId from %v", link.RemoteIdentity())
		} else {
			if mod.config.LogPings {
				mod.log.Logv(1, "ping with %v: %v", link.RemoteIdentity(), rtt)
			}
		}
	} else {
		link.Write(&frames.Ping{
			Nonce: f.Nonce,
			Pong:  true,
		})
	}
}

func (mod *Peers) handleMigrate(link *Link, f *frames.Migrate) {
	sess, ok := mod.sessions.Get(f.Nonce)
	if !ok {
		return
	}

	if sess.getState() != stateMigrating {
		return
	}

	reader, ok := sess.reader.(*muxSessionReader)
	if !ok {
		mod.log.Errorv(1, "received migrate frame on non-mux session")
		return
	}

	// Peer is done sending on the old stream. Advance closes the old buffer
	// and promotes nextBuffer immediately if already empty, or defers to the
	// Read loop on EOF otherwise.
	reader.Advance()
}

func (mod *Peers) addLink(
	link *Link,
) (err error) {
	var (
		dir     = "in"
		netName = "unknown network"
	)

	if link.outbound {
		dir = "out"
	}

	// try to figure out the network name
	switch {
	case link.LocalEndpoint() != nil:
		netName = link.LocalEndpoint().Network()
	case link.RemoteEndpoint() != nil:
		netName = link.RemoteEndpoint().Network()
	}

	err = mod.links.Add(link)
	if err != nil {
		return
	}

	// log stream addition
	mod.log.Infov(1, "added %v-stream with %v (%v)", dir, link.RemoteIdentity(), netName)
	linksWithSameIdentity := mod.links.Select(func(v *Link) bool {
		return v.RemoteIdentity().IsEqual(link.RemoteIdentity())
	})

	if !link.outbound {
		mod.linkPool.notifyLinkWatchers(link, nil)
	}

	mod.Events.Emit(&nodes.LinkCreatedEvent{
		RemoteIdentity: link.RemoteIdentity(),
		LinkID:         link.id,
		LinkCount:      len(linksWithSameIdentity),
	})

	// handle the stream
	go func() {
		mod.readLinkFrames(link)

		// remove the stream and its connections
		mod.links.Remove(link)
		for _, c := range mod.sessions.Select(func(k astral.Nonce, v *session) (ok bool) {
			return v.isOnLink(link)
		}) {
			c.Close()
		}

		linksWithSameIdentity := mod.links.Select(func(v *Link) bool {
			return v.RemoteIdentity().IsEqual(link.RemoteIdentity())
		})

		mod.Events.Emit(&nodes.LinkClosedEvent{
			RemoteIdentity: link.RemoteIdentity(),
			Forced:         false,
			LinkCount:      astral.Int8(len(linksWithSameIdentity)),
		})

		mod.log.Info("closed %v-stream with %v (%v): %v", dir, link.RemoteIdentity(), netName, link.Err())
	}()

	// reflect the stream
	go mod.reflectLink(link)

	return
}

// reflectStream reflects the observed remote endpoint on inbound streams
func (mod *Peers) reflectLink(link *Link) (err error) {
	// reflect only on inbound streams with a known remote endpoint
	if link.outbound || link.RemoteEndpoint() == nil {
		return
	}

	// note: rethink maybe switch (?)
	if _, ok := link.RemoteEndpoint().(*gateway.Endpoint); ok {
		// dont reflect gateway endpoints
		return
	}
	// reflect the endpoint
	err = mod.Objects.Push(mod.ctx, link.RemoteIdentity(),
		&nodes.ObservedEndpointMessage{
			Endpoint: link.RemoteEndpoint(),
		})

	// log the result
	if err != nil {
		mod.log.Errorv(2, "Objects.Push(%v, %v): %v", link.RemoteIdentity(), link.RemoteEndpoint(), err)
	} else {
		mod.log.Logv(2, "reflected endpoint %v to %v", link.RemoteEndpoint(), link.RemoteIdentity())
	}

	return
}

// readStreamFrames reads frames from the stream until it closes
func (mod *Peers) readLinkFrames(link *Link) {
	// read frames
	for frame := range link.Read() {
		mod.in <- &Frame{
			Frame:  frame, // NOTE: add timeout handling?
			Source: link,
		}
	}
}

// isLinked returns true if there's at least one stream with the given remote identity
func (mod *Peers) isLinked(remoteID *astral.Identity) bool {
	for _, link := range mod.links.Clone() {
		if link.RemoteIdentity().IsEqual(remoteID) {
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

func (mod *Peers) setInboundLinkNonce(aconn astral.Conn) (nonce astral.Nonce, err error) {
	nonce = astral.NewNonce()
	if _, err = nonce.WriteTo(aconn); err != nil {
		return 0, err
	}

	return nonce, nil
}

func (mod *Peers) readOutboundLinkNonce(aconn astral.Conn) (nonce astral.Nonce, err error) {
	if _, err = nonce.ReadFrom(aconn); err != nil {
		return 0, err
	}

	return nonce, nil
}

func (mod *Peers) EstablishOutboundLink(ctx context.Context, remoteID *astral.Identity, conn exonet.Conn) (_ *Link, err error) {
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

		nonce, err := mod.readOutboundLinkNonce(aconn)
		if err != nil {
			return nil, fmt.Errorf(`read outbound link nonce: %w`, err)
		}

		link := newLink(aconn, nonce, true)
		err = mod.addLink(link)

		return link, err
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
				nonce, err := mod.setInboundLinkNonce(aconn)
				if err != nil {
					return fmt.Errorf("failed to set inbound link nonce: %w", err)
				}

				link := newLink(aconn, nonce, false)
				err = mod.addLink(link)
				if err != nil {
					return fmt.Errorf("failed to add link: %w", err)
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
