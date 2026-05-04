package nodes

import (
	"context"

	"github.com/cryptopunkscc/astrald/mod/nodes"
)

type SessionMigrator struct {
	mod            *Module
	session        *session
	reader         *muxSessionReader
	writer         *muxSessionWriter
	peerBuffer     int
	oldLink        *Link
	targetLink     *Link
	oldInputBuffer *InputBuffer
}

func (mod *Module) newSessionMigrator(session *session) (*SessionMigrator, error) {
	reader, ok := session.reader.(*muxSessionReader)
	if !ok {
		return nil, nodes.ErrMigrationNotSupported
	}

	writer, ok := session.writer.(*muxSessionWriter)
	if !ok {
		return nil, nodes.ErrMigrationNotSupported
	}

	return &SessionMigrator{mod: mod, session: session, reader: reader, writer: writer}, nil
}

// Begin transitions the session to stateMigrating, stages new link buffers, and
// pauses the writer. It does not transfer mux ownership; Complete() does that.
func (m *SessionMigrator) Begin(target *Link) error {
	if !m.session.swapState(stateOpen, stateMigrating) {
		m.mod.log.Logv(1, "session %v in state %v, cannot migrate", m.session.Nonce, m.session.getState())
		return nodes.ErrInvalidSessionState
	}

	m.targetLink = target

	m.mod.log.Logv(1, "pausing session %v", m.session.Nonce)
	m.writer.Pause()

	m.oldInputBuffer = m.reader.Buf()

	newInputBuffer := NewInputBuffer(defaultBufferSize, target.Mux.sessionOnReadFunc(m.session.Nonce))
	newOutputBuffer := NewOutputBuffer(target.Mux.sessionOnWriteFunc(m.session.Nonce))
	resetFunc := func() { target.Mux.resetSession(m.session.Nonce) }

	m.writer.SwapBuf(newOutputBuffer, resetFunc)
	m.reader.SetNextBuffer(newInputBuffer)
	m.mod.log.Logv(1, "session %v migrating link %v → %v", m.session.Nonce, m.oldLink.ID(), target.ID())

	return nil
}

func (m *SessionMigrator) SendMigrateFrame() error {
	m.mod.log.Logv(1, "sending migrate frame for session %v on link %v", m.session.Nonce, m.oldLink.ID())
	return m.oldLink.Mux.migrateSession(m.session.Nonce)
}

func (m *SessionMigrator) WaitClosed(ctx context.Context) error {
	m.mod.log.Logv(1, "waiting for old input buffer to close for session %v", m.session.Nonce)
	select {
	case <-m.oldInputBuffer.Closed():
		m.mod.log.Logv(1, "old input buffer closed for session %v", m.session.Nonce)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (m *SessionMigrator) SetPeerBuffer(n int) {
	m.peerBuffer = n
}

// Complete commits the ownership transfer: removes the session from the old mux,
// inserts it into the target mux, updates session.link and session.onClose, then
// resumes traffic. On any failure the session is closed.
func (m *SessionMigrator) Complete() error {
	nonce := m.session.Nonce

	m.session.cond.L.Lock()
	if m.session.state.Load() != stateMigrating {
		m.session.cond.L.Unlock()
		return nodes.ErrInvalidSessionState
	}

	m.oldLink.Mux.sessions.Delete(nonce)

	if _, ok := m.targetLink.Mux.sessions.Set(nonce, m.session); !ok {
		m.session.cond.L.Unlock()
		m.session.Close()
		return nodes.ErrInvalidSessionState
	}

	m.session.onClose = func() { m.targetLink.Mux.sessions.Delete(nonce) }
	m.session.state.Store(stateOpen)
	m.session.cond.L.Unlock()

	m.mod.log.Logv(1, "session %v now on link %v (peer buffer %v)", nonce, m.targetLink.ID(), m.peerBuffer)

	m.writer.Grow(m.peerBuffer)
	m.writer.Resume()

	return nil
}
