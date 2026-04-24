package nodes

import (
	"context"

	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/nodes/frames"
)

type SessionMigrator struct {
	mod             *Module
	session         *session
	reader          *muxSessionReader
	writer          *muxSessionWriter
	peerBuffer      int
	oldStream       *Stream
	oldInputBuffer  *InputBuffer
	oldOutputBuffer *OutputBuffer
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

	m.session.paused = true
	for m.session.writing {
		m.session.stateCond.Wait()
	}

	if !m.session.swapState(stateOpen, stateMigrating) {
		m.session.paused = false
		m.session.stateCond.Broadcast()
		return nodes.ErrInvalidSessionState
	}

	m.oldStream = m.session.stream
	m.session.stream = target

	m.oldInputBuffer = m.reader.Buf()
	m.oldOutputBuffer = m.writer.Buf()

	newInputBuffer := m.mod.peers.newMuxInputBuffer(target, m.session.Nonce)
	newOutputBuffer := m.mod.peers.newMuxOutputBuffer(target, m.session.Nonce, m.session)

	m.writer.SetBuf(newOutputBuffer)
	m.reader.SetNextBuffer(newInputBuffer)

	return nil
}

func (m *SessionMigrator) Rollback() {
	m.session.stateCond.L.Lock()
	defer m.session.stateCond.L.Unlock()

	m.writer.SetBuf(m.oldOutputBuffer)
	m.reader.SetNextBuffer(nil)
	m.session.stream = m.oldStream
	m.session.swapState(stateMigrating, stateOpen)
	m.session.Resume()
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

func (m *SessionMigrator) Resume() {
	m.writer.Grow(m.peerBuffer)
	m.session.swapState(stateMigrating, stateOpen)
	m.session.Resume()
}
