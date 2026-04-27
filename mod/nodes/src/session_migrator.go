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
	oldStream      *Stream
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

func (m *SessionMigrator) BeginMigrate(target *Stream) error {
	m.mod.log.Logv(1, "pausing session %v", m.session.Nonce)
	m.writer.Pause()

	m.session.cond.L.Lock()
	if m.session.closed {
		m.session.cond.L.Unlock()
		m.writer.Resume()
		m.mod.log.Logv(1, "session %v closed before migration could begin", m.session.Nonce)
		return nodes.ErrInvalidSessionState
	}
	m.oldStream = m.session.stream
	m.session.stream = target
	m.session.state = stateMigrating
	m.session.cond.L.Unlock()

	m.oldInputBuffer = m.reader.Buf()

	newInputBuffer := m.mod.peers.newMuxInputBuffer(target, m.session.Nonce)
	newOutputBuffer := m.mod.peers.newMuxOutputBuffer(target, m.session.Nonce, m.session)

	m.writer.SetBuf(newOutputBuffer)
	m.reader.SetNextBuffer(newInputBuffer)
	m.mod.log.Logv(1, "session %v migrating stream %v → %v", m.session.Nonce, m.oldStream.id, target.id)

	return nil
}

func (m *SessionMigrator) SendMigrateFrame() error {
	m.mod.log.Logv(1, "sending migrate frame for session %v on stream %v", m.session.Nonce, m.oldStream.id)
	return m.oldStream.Write(&frames.Migrate{Nonce: m.session.Nonce})
}

func (m *SessionMigrator) WaitDrain(ctx context.Context) error {
	if m.oldInputBuffer.IsEmpty() {
		return nil
	}

	m.mod.log.Logv(1, "waiting for old input buffer to drain for session %v", m.session.Nonce)
	select {
	case <-m.oldInputBuffer.Done():
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (m *SessionMigrator) SetPeerBuffer(n int) {
	m.peerBuffer = n
}

func (m *SessionMigrator) Resume() error {
	m.mod.log.Logv(1, "resuming session %v on stream %v (peer buffer %v)", m.session.Nonce, m.session.stream.id, m.peerBuffer)
	m.session.setState(stateOpen)
	m.writer.Grow(m.peerBuffer)
	m.writer.Resume()
	return nil
}
