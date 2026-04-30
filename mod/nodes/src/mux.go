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

	in chan frames.Frame
}

func NewMux(ch *channel.Channel, mod *Module, link *Link) *Mux {
	return &Mux{
		mod:  mod,
		link: link,
		ch:   ch,
		in:   make(chan frames.Frame, 32),
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
	conn, ok := m.createSession(q.Nonce)
	if !ok {
		return query.RouteNotFound()
	}

	conn.RemoteIdentity = q.Target
	conn.Query = q.QueryString
	conn.Outbound = true

	if q.Caller.IsEqual(ctx.Identity()) {
		err = m.link.Send(&frames.Query{
			Nonce:  q.Nonce,
			Query:  q.QueryString,
			Buffer: uint32(defaultBufferSize),
		})
	} else {
		// TODO: reconsider relayID ownership — mux sets it here to satisfy handleResponse validation
		conn.relayID = m.link.RemoteIdentity()
		err = m.link.Send(&frames.RelayQuery{
			CallerID: q.Caller,
			TargetID: q.Target,
			Query: frames.Query{
				Nonce:  q.Nonce,
				Query:  q.QueryString,
				Buffer: uint32(defaultBufferSize),
			},
		})
	}
	if err != nil {
		conn.Close()
		return query.RouteNotFound()
	}

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

// --- session management ---

func (m *Mux) getSession(nonce astral.Nonce) (*session, bool) {
	session, ok := m.sessions.Get(nonce)
	if !ok {
		m.resetSession(nonce)
		return nil, false
	}

	return session, true
}

func (m *Mux) createSession(nonce astral.Nonce) (*session, bool) {
	session, ok := m.sessions.Set(nonce, newSession(nonce))
	if !ok {
		return nil, false
	}

	session.remove = func() { m.sessions.Delete(nonce) }
	return session, true
}

func (m *Mux) resetSession(nonce astral.Nonce) error {
	return m.link.Send(&frames.Reset{Nonce: nonce})
}

func (m *Mux) migrateSession(nonce astral.Nonce) error {
	return m.link.Send(&frames.Migrate{Nonce: nonce})
}

func (m *Mux) newInputBuffer(nonce astral.Nonce) *InputBuffer {
	onRead := func(n int) {
		m.link.Send(&frames.Read{Nonce: nonce, Len: uint32(n)})
	}

	return NewInputBuffer(defaultBufferSize, onRead)
}

func (m *Mux) newOutputBuffer(nonce astral.Nonce) *OutputBuffer {
	onWrite := func(p []byte) error {
		remaining := p
		for len(remaining) > 0 {
			chunkSize := maxPayloadSize
			if len(remaining) < chunkSize {
				chunkSize = len(remaining)
			}
			if err := m.link.Send(&frames.Data{
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

func (m *Mux) setupSession(session *session, peerBuffer int) error {
	nonce := session.Nonce
	resetFunc := func() { m.resetSession(nonce) }
	reader := newSessionReader(m.newInputBuffer(nonce))
	writer := newSessionWriter(m.newOutputBuffer(nonce), resetFunc)
	writer.Grow(peerBuffer)
	if err := session.Setup(m.link, reader, writer); err != nil {
		return err
	}
	session.Open()
	return nil
}

func (m *Mux) handleQuery(f *frames.Query) {
	m.handleInboundQuery(f.Nonce, m.link.RemoteIdentity(), m.link.LocalIdentity(), nil, f.Query, int(f.Buffer))
}

func (m *Mux) handleRelayQuery(f *frames.RelayQuery) {
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
	m.handleInboundQuery(f.Query.Nonce, f.CallerID, f.TargetID, m.link.RemoteIdentity(), f.Query.Query, int(f.Query.Buffer))
}

func (m *Mux) handleInboundQuery(nonce astral.Nonce, caller, target *astral.Identity, relayID *astral.Identity, queryStr string, initBuffer int) {
	conn, ok := m.createSession(nonce)
	if !ok {
		return
	}

	conn.RemoteIdentity = caller
	conn.relayID = relayID
	conn.Query = queryStr
	conn.stream = m.link

	q := astral.Launch(&astral.Query{
		Nonce:       nonce,
		Caller:      caller,
		Target:      target,
		QueryString: queryStr,
	})
	q.Extra.Set("origin", astral.OriginNetwork)
	ctx := astral.NewContext(nil).WithIdentity(m.mod.node.Identity())

	w, err := m.mod.node.RouteQuery(ctx, q, conn)
	if err != nil {
		conn.Close()
		var code = uint8(frames.CodeRejected)
		var reject *astral.ErrRejected
		if errors.As(err, &reject) {
			code = reject.Code
		}
		m.link.Send(&frames.Response{Nonce: nonce, ErrCode: code})
		return
	}

	if err := m.setupSession(conn, initBuffer); err != nil {
		return
	}
	m.link.Send(&frames.Response{Nonce: nonce, ErrCode: frames.CodeAccepted, Buffer: uint32(defaultBufferSize)})

	go func() {
		io.Copy(w, conn)
		w.Close()
	}()
}

func (m *Mux) handleResponse(f *frames.Response) {
	conn, ok := m.sessions.Get(f.Nonce)
	if !ok {
		return
	}

	expectedSource := conn.RemoteIdentity
	if conn.relayID != nil {
		expectedSource = conn.relayID
	}
	if !expectedSource.IsEqual(m.link.RemoteIdentity()) {
		return
	}

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

	if err := m.setupSession(conn, int(f.Buffer)); err != nil {
		return
	}
	conn.res <- 0
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

// --- frame I/O ---

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
	return m.link.Send(&frames.Ping{Nonce: nonce})
}
