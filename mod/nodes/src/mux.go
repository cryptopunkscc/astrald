package nodes

import (
	"errors"
	"fmt"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/nodes/frames"
	"github.com/cryptopunkscc/astrald/sig"
)

var _ astral.Router = &Mux{}

type Mux struct {
	mod  *Module
	link *Link
	ch   *channel.Channel

	sessions sig.Map[astral.Nonce, *session]
	in       chan frames.Frame
}

func (mod *Module) newMux(ch *channel.Channel) *Mux {
	return &Mux{
		mod: mod,
		ch:  ch,
		in:  make(chan frames.Frame, 32),
	}
}

// Run starts the mux runtime and blocks until the link dies or ctx is cancelled.
func (m *Mux) Run(ctx *astral.Context) {
	go m.reader()
	go m.link.pingLoop()

	for {
		select {
		case frame, ok := <-m.in:
			if !ok {
				m.closeAllSessions()
				return
			}

			switch f := frame.(type) {
			case *frames.Query:
				go m.handleQuery(f)
			case *frames.RelayQuery:
				go m.handleRelayQuery(f)
			case *frames.Response:
				m.handleResponse(f)
			case *frames.Ping:
				m.handlePing(f)
			case *frames.Data:
				m.handleData(f)
			case *frames.Reset:
				m.handleReset(f)
			case *frames.Read:
				m.handleRead(f)
			case *frames.Migrate:
				m.handleMigrate(f)
			default:
				m.mod.log.Errorv(2, "unknown frame: %v", frame)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (m *Mux) closeAllSessions() {
	for _, s := range m.sessions.Clone() {
		s.Close()
	}
}

// RouteQuery sends a query over this link and wires the resulting session.
func (m *Mux) RouteQuery(ctx *astral.Context, q *astral.InFlightQuery, w io.WriteCloser) (_ io.WriteCloser, err error) {
	conn, ok := m.createSession(q.Nonce, q.Target, m.link.RemoteIdentity(), q.QueryString, true, 0)
	if !ok {
		return query.RouteNotFound()
	}

	fq := frames.Query{
		Nonce:  q.Nonce,
		Query:  q.QueryString,
		Buffer: uint32(defaultBufferSize),
	}

	var frame frames.Frame = &fq
	if !q.Caller.IsEqual(ctx.Identity()) {
		frame = &frames.RelayQuery{
			CallerID: q.Caller,
			TargetID: q.Target,
			Query:    fq,
		}
	}

	err = m.link.Send(frame)
	if err != nil {
		conn.Close()
		return query.RouteNotFound()
	}

	errCode, err := sig.Recv(ctx, conn.res)
	if err != nil {
		conn.Close()
		return query.RouteNotFound()
	}

	if errCode != 0 {
		return query.RejectWithCode(errCode)
	}

	go func() {
		io.Copy(w, conn)
		w.Close()
	}()
	return conn, nil
}

func (m *Mux) getSession(nonce astral.Nonce) (*session, bool) {
	session, ok := m.sessions.Get(nonce)
	if !ok {
		m.resetSession(nonce)
		return nil, false
	}

	return session, true
}

func (m *Mux) createSession(nonce astral.Nonce, remoteIdentity, sourceIdentity *astral.Identity, queryStr string, outbound bool, peerBuffer int) (*session, bool) {
	session, ok := m.sessions.Set(nonce, newSession(nonce, remoteIdentity, sourceIdentity, queryStr, outbound))
	if !ok {
		return nil, false
	}

	session.remove = func() { m.sessions.Delete(nonce) } // todo: basically onClose hook

	resetFunc := func() { m.resetSession(nonce) }
	reader := newSessionReader(m.newInputBuffer(nonce))
	writer := newSessionWriter(m.newOutputBuffer(nonce), resetFunc)
	writer.Grow(peerBuffer)

	if err := session.Setup(m.link, reader, writer); err != nil {
		m.sessions.Delete(nonce)
		return nil, false
	}

	return session, true
}

func (m *Mux) resetSession(nonce astral.Nonce) error {
	return m.ch.Send(&frames.Reset{Nonce: nonce})
}

func (m *Mux) migrateSession(nonce astral.Nonce) error {
	return m.ch.Send(&frames.Migrate{Nonce: nonce})
}

func (m *Mux) newInputBuffer(nonce astral.Nonce) *InputBuffer {
	onRead := func(n int) {
		m.ch.Send(&frames.Read{Nonce: nonce, Len: uint32(n)})
	}

	return NewInputBuffer(defaultBufferSize, onRead)
}

func (m *Mux) newOutputBuffer(nonce astral.Nonce) *OutputBuffer {
	onWrite := func(p []byte) error {
		remaining := p
		for len(remaining) > 0 {
			chunkSize := maxPayloadSize // configurable
			if len(remaining) < chunkSize {
				chunkSize = len(remaining)
			}

			if err := m.ch.Send(&frames.Data{
				Nonce:   nonce,
				Payload: remaining[:chunkSize],
			}); err != nil {
				return err
			}

			m.link.AddThroughputBytes(chunkSize)
			remaining = remaining[chunkSize:]
		}

		return nil
	}

	return NewOutputBuffer(onWrite)
}

func (m *Mux) handleQuery(f *frames.Query) {
	m.handleInboundQuery(f.Nonce, m.link.RemoteIdentity(), m.link.LocalIdentity(), f.Query, int(f.Buffer))
}

func (m *Mux) handleRelayQuery(f *frames.RelayQuery) {
	// caller is relaying a query
	if !f.CallerID.IsEqual(m.link.RemoteIdentity()) {

		ctx := astral.NewContext(nil).WithIdentity(m.mod.node.Identity())
		if !m.mod.Auth.Authorize(ctx, &nodes.RelayForAction{
			Action: auth.NewAction(m.link.RemoteIdentity()),
			ForID:  f.CallerID,
		}) {
			m.link.Send(&frames.Response{Nonce: f.Query.Nonce, ErrCode: frames.CodeRejected})
			return
		}
	}

	m.handleInboundQuery(f.Query.Nonce, f.CallerID, f.TargetID, f.Query.Query, int(f.Query.Buffer))
}

// caller is the logical remote identity; m.link.RemoteIdentity() is the transport source
func (m *Mux) handleInboundQuery(nonce astral.Nonce, caller, target *astral.Identity, queryStr string, initBuffer int) {
	session, ok := m.createSession(nonce, caller, m.link.RemoteIdentity(), queryStr, false, initBuffer)
	if !ok {
		return
	}

	q := astral.Launch(&astral.Query{
		Nonce:       nonce,
		Caller:      caller,
		Target:      target,
		QueryString: queryStr,
	})
	q.Extra.Set("origin", astral.OriginNetwork)
	ctx := astral.NewContext(nil).WithIdentity(m.mod.node.Identity())

	w, err := m.mod.node.RouteQuery(ctx, q, session)
	if err != nil {
		session.Close()
		var code = uint8(frames.CodeRejected)
		var reject *astral.ErrRejected
		if errors.As(err, &reject) {
			code = reject.Code
		}
		m.link.Send(&frames.Response{Nonce: nonce, ErrCode: code})
		return
	}

	m.link.Send(&frames.Response{Nonce: nonce, ErrCode: frames.CodeAccepted, Buffer: uint32(defaultBufferSize)})
	session.Open()

	go func() {
		io.Copy(w, session)
		w.Close()
	}()
}

func (m *Mux) handleResponse(f *frames.Response) {
	session, ok := m.sessions.Get(f.Nonce)
	if !ok {
		return
	}

	if !session.acceptsSource(m.link.RemoteIdentity()) {
		return
	}

	if !session.rejectRoute(f.ErrCode) {
		return
	}

	if session.getState() != stateRouting {
		return
	}

	writer, ok := session.writer.(*muxSessionWriter)
	if !ok {
		m.mod.log.Errorv(1, "received response for session %v without mux writer", f.Nonce)
		session.Close()
		return
	}

	writer.Grow(int(f.Buffer))
	session.Open()
	session.res <- 0
}

func (m *Mux) handleData(f *frames.Data) {
	session, ok := m.getSession(f.Nonce)
	if !ok {
		return
	}

	switch session.getState() {
	case stateOpen, stateMigrating:
	default:
		m.mod.log.Errorv(1, "received data frame from %v in state %v", m.link.RemoteIdentity(), session.getState())
		m.link.Send(&frames.Reset{Nonce: f.Nonce})
		return
	}

	m.link.AddThroughputBytes(len(f.Payload))

	reader, ok := session.reader.(*muxSessionReader)
	if !ok {
		m.mod.log.Errorv(1, "received data frame from %v on non-mux session", m.link.RemoteIdentity())
		return
	}

	if err := reader.Push(f.Payload); err != nil {
		m.mod.log.Errorv(1, "failed to push read frame: %v", err)
		session.Close()
	}
}

func (m *Mux) handleRead(f *frames.Read) {
	session, ok := m.getSession(f.Nonce)
	if !ok {
		return
	}

	if !session.isOnStream(m.link) {
		return
	}

	if writer, ok := session.writer.(*muxSessionWriter); ok {
		writer.Grow(int(f.Len))
	}
}

func (m *Mux) handleReset(f *frames.Reset) {
	session, ok := m.sessions.Get(f.Nonce)
	if !ok {
		return
	}

	if w, ok := session.writer.(*muxSessionWriter); ok {
		w.PeerClose()
	}

	session.Close()
}

func (m *Mux) handlePing(f *frames.Ping) {
	if f.Pong {
		rtt, err := m.link.pong(f.Nonce)
		if err != nil {
			m.mod.log.Errorv(1, "invalid pong nonce from %v", m.link.RemoteIdentity())
		} else if m.mod.config.LogPings {
			m.mod.log.Logv(1, "ping with %v: %v", m.link.RemoteIdentity(), rtt)
		}
	} else {
		m.link.Send(&frames.Ping{Nonce: f.Nonce, Pong: true})
	}
}

func (m *Mux) handleMigrate(f *frames.Migrate) {
	session, ok := m.sessions.Get(f.Nonce)
	if !ok {
		return
	}
	if session.getState() != stateMigrating {
		return
	}
	reader, ok := session.reader.(*muxSessionReader)
	if !ok {
		m.mod.log.Errorv(1, "received migrate frame on non-mux session")
		return
	}

	reader.Advance()
}

func (m *Mux) reader() {
	var rerr error
	defer func() {
		m.link.CloseWithError(rerr)
		close(m.in)
		close(m.link.done)
	}()

	for {
		obj, err := m.link.Receive()
		if err != nil {
			rerr = err
			return
		}

		frame, ok := obj.(frames.Frame)
		if !ok {
			rerr = fmt.Errorf("decoded object is not a Frame: %T", obj)
			return
		}

		m.in <- frame
	}
}

func (m *Mux) Read() <-chan frames.Frame { return m.in }

func (m *Mux) ping(nonce astral.Nonce) error {
	return m.ch.Send(&frames.Ping{Nonce: nonce})
}
