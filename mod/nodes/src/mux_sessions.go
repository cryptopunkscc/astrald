package nodes

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nodes/frames"
)

func (m *Mux) createSession(nonce astral.Nonce, remoteIdentity, sourceIdentity *astral.Identity, queryStr string, outbound bool, peerBuffer int) (*session, bool) {
	session, ok := m.sessions.Set(nonce, newSession(nonce, remoteIdentity, sourceIdentity, queryStr, outbound))
	if !ok {
		return nil, false
	}

	session.onClose = m.sessionOnCloseFunc(nonce)

	reader := newSessionReader(NewInputBuffer(defaultBufferSize, m.sessionOnReadFunc(nonce)))
	writer := newSessionWriter(NewOutputBuffer(m.sessionOnWriteFunc(nonce)), m.sessionResetFunc(nonce))
	writer.Grow(peerBuffer)

	if err := session.Setup(reader, writer); err != nil {
		m.sessions.Delete(nonce)
		return nil, false
	}

	return session, true
}

func (m *Mux) closeAllSessions() {
	for _, s := range m.sessions.Clone() {
		s.Close()
	}
}

func (m *Mux) sessionOnCloseFunc(nonce astral.Nonce) func() {
	return func() { m.sessions.Delete(nonce) }
}

func (m *Mux) sessionResetFunc(nonce astral.Nonce) func() {
	return func() { m.resetSession(nonce) }
}

func (m *Mux) sessionOnReadFunc(nonce astral.Nonce) func(int) {
	return func(n int) {
		m.ch.Send(&frames.Read{Nonce: nonce, Len: uint32(n)})
	}
}

func (m *Mux) sessionOnWriteFunc(nonce astral.Nonce) func([]byte) error {
	return func(p []byte) error {
		remaining := p
		for len(remaining) > 0 {
			chunkSize := maxPayloadSize
			if len(remaining) < chunkSize {
				chunkSize = len(remaining)
			}

			if err := m.ch.Send(&frames.Data{
				Nonce:   nonce,
				Payload: remaining[:chunkSize],
			}); err != nil {
				return err
			}

			m.addBytes(chunkSize)
			remaining = remaining[chunkSize:]
		}

		return nil
	}
}
