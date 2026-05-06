package nodes

import (
	"context"

	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/nodes/frames"
)

type SessionMigrator struct {
	mod            *Module
	session        *session
	reader         *muxSessionReader
	writer         *muxSessionWriter
	peerBuffer     int
	oldLink        *Link
	oldMux         *Mux
	newMux         *Mux
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
	m.oldLink = m.session.link
	m.session.link = target
	m.session.cond.L.Unlock()
	m.oldMux = m.oldLink.GetMux()
	m.newMux = target.GetMux()

	m.mod.log.Logv(1, "pausing session %v", m.session.Nonce)
	m.writer.Pause()

	m.oldInputBuffer = m.reader.Buf()

	newInputBuffer := NewInputBuffer(defaultBufferSize, m.newMux.sessionOnReadFunc(m.session.Nonce))
	newOutputBuffer := NewOutputBuffer(m.newMux.sessionOnWriteFunc(m.session.Nonce))
	resetFunc := m.newMux.sessionResetFunc(m.session.Nonce)

	m.writer.SwapBuf(newOutputBuffer, resetFunc)
	m.reader.SetNextBuffer(newInputBuffer)
	m.mod.log.Logv(1, "session %v migrating link %v → %v", m.session.Nonce, m.oldLink.id, target.id)

	return nil
}

func (m *SessionMigrator) SendMigrateFrame() error {
	m.mod.log.Logv(1, "sending migrate frame for session %v on link %v", m.session.Nonce, m.oldLink.id)
	return m.oldLink.Stream.Write(&frames.Migrate{Nonce: m.session.Nonce})
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
	m.mod.log.Logv(1, "resuming session %v on link %v (peer buffer %v)", m.session.Nonce, m.session.link.id, m.peerBuffer)

	if m.oldMux != nil && m.newMux != nil && !m.oldMux.transferSessionTo(m.newMux, m.session) {
		return nodes.ErrInvalidSessionState
	}

	m.session.setState(stateOpen)
	m.writer.Grow(m.peerBuffer)
	m.writer.Resume()
	return nil
}
