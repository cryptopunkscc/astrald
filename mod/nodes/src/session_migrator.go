package nodes

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nodes/frames"
)

type sessionMigrator struct {
	mod       *Module
	session   *session
	target    *Stream
	oldStream *Stream
	oldBuffer *InputBuffer
}

func (mod *Module) newSessionMigrator(sess *session, targetStream *Stream) *sessionMigrator {
	return &sessionMigrator{
		mod:     mod,
		session: sess,
		target:  targetStream,
	}
}

// prepare pauses the writer, swaps buffers and stream to the target.
func (m *sessionMigrator) prepare() {

}

// Migrate prepares the session for migration and writes the Migrate frame on the old stream,
// telling the remote side that no more Data frames will follow for this session.
func (m *sessionMigrator) Migrate() error {
	newInBuf, newOutBuf := m.mod.peers.newSessionBuffers(m.target, m.session.Nonce, m.session)

	m.oldBuffer = m.session.reader.CurrentBuffer()
	m.oldStream = m.session.stream

	m.session.stream = m.target
	m.session.swapState(stateOpen, stateMigrating)
	m.session.writer.Pause()

	m.session.reader.SetNextBuffer(newInBuf)
	m.session.writer.SetBuf(newOutBuf)

	err := m.oldStream.Write(&frames.Migrate{Nonce: m.session.Nonce})
	if err != nil {
		m.session.writer.Resume()
		m.session.reader.SetNextBuffer(nil)
		m.session.swapState(stateMigrating, stateOpen)
	}

	return nil
}

// Drain waits for the old input buffer to be closed (by handleMigrate) and drained.
func (m *sessionMigrator) Drain(ctx *astral.Context) error {
	select {
	case <-m.oldBuffer.Done():
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Resume grants write credit, unpauses the writer, and returns session to open state.
func (m *sessionMigrator) Resume(credit int) {
	m.session.writer.Grow(credit)
	m.session.writer.Resume()
	m.session.swapState(stateMigrating, stateOpen)
}
