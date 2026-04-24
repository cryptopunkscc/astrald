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
	m.session.stateCond.L.Lock()
	defer m.session.stateCond.L.Unlock()

	m.writer.Pause()

	if !m.session.swapState(stateOpen, stateMigrating) {
		m.writer.Resume()
		return nodes.ErrInvalidSessionState
	}

	m.oldStream = m.session.stream
	m.session.stream = target

	m.oldInputBuffer = m.reader.Buf()

	newInputBuffer := m.mod.peers.newMuxInputBuffer(target, m.session.Nonce)
	newOutputBuffer := m.mod.peers.newMuxOutputBuffer(target, m.session.Nonce, m.session)

	m.writer.SetBuf(newOutputBuffer)
	m.reader.SetNextBuffer(newInputBuffer)

	return nil
}

func (m *SessionMigrator) SendMigrateFrame() error {
	// note: will be part of mux
	return m.oldStream.Write(&frames.Migrate{Nonce: m.session.Nonce})
}

func (m *SessionMigrator) WaitDrain(ctx context.Context) error {
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
	m.session.stateCond.L.Lock()
	defer m.session.stateCond.L.Unlock()
	if !m.session.swapState(stateMigrating, stateOpen) {
		return nodes.ErrInvalidSessionState
	}

	m.writer.Grow(m.peerBuffer)
	m.writer.Resume()
	return nil
}
