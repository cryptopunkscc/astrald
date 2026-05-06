package nodes

import (
	"errors"
	"io"
	"sync"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/nodes/frames"
	"github.com/cryptopunkscc/astrald/sig"
)

type Mux struct {
	mod  *Module
	link *Link

	sessions sig.Map[astral.Nonce, *session]

	cond   *sync.Cond
	router astral.Router
}

func newMux(mod *Module, link *Link) *Mux {
	m := &Mux{
		mod:  mod,
		link: link,
	}
	m.cond = sync.NewCond(&sync.Mutex{})
	return m
}

func (m *Mux) SetRouter(r astral.Router) {
	m.cond.L.Lock()
	defer m.cond.L.Unlock()
	if m.router != nil {
		return
	}
	m.router = r
	m.cond.Broadcast()
}

func (m *Mux) RouteQuery(ctx *astral.Context, q *astral.InFlightQuery, w io.WriteCloser) (_ io.WriteCloser, err error) {
	sourceID := m.link.RemoteIdentity()
	if q.Caller.IsEqual(ctx.Identity()) {
		sourceID = nil
	}

	conn, ok := m.createSession(q.Nonce, q.Target, sourceID, q.QueryString, true, 0)
	if !ok {
		return query.RouteNotFound()
	}

	queryFrame := frames.Query{
		Nonce:  q.Nonce,
		Query:  q.QueryString,
		Buffer: uint32(defaultBufferSize),
	}

	var frame frames.Frame = &queryFrame
	if !q.Caller.IsEqual(ctx.Identity()) {
		frame = &frames.RelayQuery{
			CallerID: q.Caller,
			TargetID: q.Target,
			Query:    queryFrame,
		}
	}

	if err := m.link.Stream.Write(frame); err != nil {
		conn.Close()
		return nil, err
	}

	select {
	case errCode := <-conn.routingResult:
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

func (m *Mux) HandleFrame(frame frames.Frame) {
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
}

func (m *Mux) handleQuery(f *frames.Query) {
	m.handleInboundQuery(f.Nonce, m.link.RemoteIdentity(), m.link.LocalIdentity(), nil, f.Query, int(f.Buffer))
}

func (m *Mux) handleRelayQuery(relayQuery *frames.RelayQuery) error {
	if !relayQuery.CallerID.IsEqual(m.link.RemoteIdentity()) {
		ctx := astral.NewContext(nil).WithIdentity(m.mod.node.Identity())
		if !m.mod.Auth.Authorize(ctx, &nodes.RelayForAction{
			Action: auth.NewAction(m.link.RemoteIdentity()),
			ForID:  relayQuery.CallerID,
		}) {
			_ = m.link.Stream.Write(&frames.Response{Nonce: relayQuery.Query.Nonce, ErrCode: frames.CodeRejected})
			return nil
		}
	}

	m.handleInboundQuery(
		relayQuery.Query.Nonce,
		relayQuery.CallerID,
		relayQuery.TargetID,
		m.link.RemoteIdentity(),
		relayQuery.Query.Query,
		int(relayQuery.Query.Buffer),
	)

	return nil
}

func (m *Mux) handleInboundQuery(linkNonce astral.Nonce, caller, target, relayID *astral.Identity, queryStr string, initBuffer int) {
	conn, ok := m.createSession(linkNonce, caller, relayID, queryStr, false, initBuffer)
	if !ok {
		return
	}

	ctx := astral.NewContext(nil).WithIdentity(m.mod.node.Identity())

	router, err := m.waitRouter(ctx)
	if err != nil {
		conn.Close()
		m.link.Stream.Write(&frames.Response{Nonce: linkNonce, ErrCode: frames.CodeRejected})
		return
	}

	q := astral.Launch(&astral.Query{
		Nonce:       linkNonce,
		Caller:      caller,
		Target:      target,
		QueryString: queryStr,
	})

	q.Extra.Set("origin", astral.OriginNetwork)
	w, err := router.RouteQuery(ctx, q, conn)
	if err != nil {
		conn.Close()
		code := uint8(frames.CodeRejected)
		var reject *astral.ErrRejected
		if errors.As(err, &reject) {
			code = reject.Code
		}
		m.link.Stream.Write(&frames.Response{Nonce: linkNonce, ErrCode: code})
		return
	}

	conn.setState(stateOpen)
	m.link.Stream.Write(&frames.Response{Nonce: linkNonce, ErrCode: frames.CodeAccepted, Buffer: uint32(defaultBufferSize)})
	conn.Open()

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

	if conn.SourceIdentity != nil && !conn.SourceIdentity.IsEqual(m.link.RemoteIdentity()) {
		return
	}

	if conn.SourceIdentity == nil && !conn.RemoteIdentity.IsEqual(m.link.RemoteIdentity()) {
		return
	}

	if f.ErrCode != 0 {
		if !conn.swapState(stateRouting, stateClosed) {
			return
		}
		conn.routingResult <- f.ErrCode
		return
	}

	if !conn.swapState(stateRouting, stateOpen) {
		return
	}

	if writer, ok := conn.writer.(*muxSessionWriter); ok {
		writer.Grow(int(f.Buffer))
	}
	conn.Open()

	conn.routingResult <- 0
}

func (m *Mux) handleData(f *frames.Data) {
	session, ok := m.sessions.Get(f.Nonce)
	if !ok {
		m.link.Stream.Write(&frames.Reset{Nonce: f.Nonce})
		return
	}

	switch session.getState() {
	case stateOpen, stateMigrating:
	default:
		m.mod.log.Errorv(1, "received data frame from %v in state %v", m.link.RemoteIdentity(), session.getState())
		m.link.Stream.Write(&frames.Reset{Nonce: f.Nonce})
		return
	}

	m.addBytes(len(f.Payload))

	reader, ok := session.reader.(*muxSessionReader)
	if !ok {
		m.mod.log.Errorv(1, "received data frame from %v on non-mux session", m.link.RemoteIdentity())
		return
	}

	if err := reader.Push(f.Payload); err != nil {
		m.mod.log.Errorv(1, "failed to push read frame: %v", err)
		session.Close()
		return
	}
}

func (m *Mux) handleRead(f *frames.Read) {
	session, ok := m.sessions.Get(f.Nonce)
	if !ok {
		m.link.Stream.Write(&frames.Reset{Nonce: f.Nonce})
		return
	}

	if !session.isOnLink(m.link) {
		return
	}

	writer, ok := session.writer.(*muxSessionWriter)
	if ok {
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
			m.mod.log.Errorv(1, "invalid pong sessionId from %v", m.link.RemoteIdentity())
		} else if m.mod.config.LogPings {
			m.mod.log.Logv(1, "ping with %v: %v", m.link.RemoteIdentity(), rtt)
		}
		return
	}

	m.link.Stream.Write(&frames.Ping{
		Nonce: f.Nonce,
		Pong:  true,
	})
}

func (m *Mux) handleMigrate(f *frames.Migrate) {
	sess, ok := m.sessions.Get(f.Nonce)
	if !ok {
		return
	}

	if sess.getState() != stateMigrating {
		return
	}

	reader, ok := sess.reader.(*muxSessionReader)
	if !ok {
		m.mod.log.Errorv(1, "received migrate frame on non-mux session")
		return
	}

	reader.Advance()
}

func (m *Mux) addBytes(n int) {
	m.link.throughput.Add(uint64(n))
	if m.link.pressure != nil {
		m.link.pressure.OnBytes(n, time.Now())
	}
}

// waitRouter blocks until the mux has a router or ctx is cancelled.
func (m *Mux) waitRouter(ctx *astral.Context) (astral.Router, error) {
	stop := make(chan struct{})
	defer close(stop)
	go func() {
		select {
		case <-ctx.Done():
			m.cond.Broadcast()
		case <-stop:
		}
	}()

	m.cond.L.Lock()
	defer m.cond.L.Unlock()
	for m.router == nil {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		m.cond.Wait()
	}
	return m.router, nil
}
