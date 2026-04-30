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
	oldStream      *Link
	oldInputBuffer *InputBuffer
}

func (mod *Module) newSessionMigrator(sess *session) (*SessionMigrator, error) {
	reader, ok := sess.reader.(*muxSessionReader)
	if !ok {
		return nil, nodes.ErrMigrationNotSupported
	}

	writer, ok := sess.writer.(*muxSessionWriter)
	if !ok {
		return nil, nodes.ErrMigrationNotSupported
	}

	return &SessionMigrator{mod: mod, session: sess, reader: reader, writer: writer}, nil
}

func (m *SessionMigrator) Begin(target *Link) error {
	if !m.session.swapState(stateOpen, stateMigrating) {
		m.mod.log.Logv(1, "session %v in state %v, cannot migrate", m.session.Nonce, m.session.getState())
		return nodes.ErrInvalidSessionState
	}
	m.session.cond.L.Lock()
	m.oldStream = m.session.stream
	m.session.stream = target
	m.session.cond.L.Unlock()

	m.mod.log.Logv(1, "pausing session %v", m.session.Nonce)
	m.writer.Pause()

	m.oldInputBuffer = m.reader.Buf()

	newInputBuffer := target.Mux.newInputBuffer(m.session.Nonce)
	newOutputBuffer := target.Mux.newOutputBuffer(m.session.Nonce)
	resetFunc := func() { target.Mux.resetSession(m.session.Nonce) }

	m.writer.SwapBuf(newOutputBuffer, resetFunc)
	m.reader.SetNextBuffer(newInputBuffer)
	m.mod.log.Logv(1, "session %v migrating stream %v → %v", m.session.Nonce, m.oldStream.id, target.id)

	return nil
}

func (m *SessionMigrator) SendMigrateFrame() error {
	m.mod.log.Logv(1, "sending migrate frame for session %v on stream %v", m.session.Nonce, m.oldStream.id)
	return m.oldStream.Mux.migrateSession(m.session.Nonce)
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

func (m *SessionMigrator) Complete() error {
	m.mod.log.Logv(1, "resuming session %v on stream %v (peer buffer %v)", m.session.Nonce, m.session.stream.id, m.peerBuffer)

	m.session.setState(stateOpen)
	m.writer.Grow(m.peerBuffer)
	m.writer.Resume()
	return nil
}
