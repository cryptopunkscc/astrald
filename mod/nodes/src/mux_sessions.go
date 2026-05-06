package nodes

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nodes/frames"
)

func (mux *Mux) adoptSession(sess *session) bool {
	_, ok := mux.sessions.Set(sess.Nonce, sess)
	if !ok {
		return false
	}
	sess.onClose = func() { mux.sessions.Delete(sess.Nonce) }
	return ok
}

func (mux *Mux) transferSessionTo(target *Mux, sess *session) bool {
	if target == nil || target == mux {
		return target == mux
	}

	if !target.adoptSession(sess) {
		return false
	}

	mux.sessions.Delete(sess.Nonce)
	return true
}

func (mux *Mux) getSession(nonce astral.Nonce) (*session, bool) {
	return mux.sessions.Get(nonce)
}

func (mux *Mux) sessionsList() []*session {
	return mux.sessions.Values()
}

func (m *Mux) createSession(nonce astral.Nonce, remoteIdentity, sourceIdentity *astral.Identity, queryStr string, outbound bool, peerBuffer int) (*session, bool) {
	session, ok := m.sessions.Set(nonce, newSession(nonce, remoteIdentity, sourceIdentity, queryStr, outbound))
	if !ok {
		return nil, false
	}

	session.onClose = m.sessionOnCloseFunc(nonce)

	reader := newSessionReader(NewInputBuffer(defaultBufferSize, m.sessionOnReadFunc(nonce)))
	writer := newSessionWriter(NewOutputBuffer(m.sessionOnWriteFunc(nonce)), m.sessionResetFunc(nonce))
	writer.Grow(peerBuffer)

	if err := session.Setup(m.link, reader, writer); err != nil {
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
	return func() { m.link.Stream.Write(&frames.Reset{Nonce: nonce}) }
}

func (m *Mux) sessionOnReadFunc(nonce astral.Nonce) func(int) {
	return func(n int) {
		m.link.Stream.Write(&frames.Read{Nonce: nonce, Len: uint32(n)})
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

			if err := m.link.Stream.Write(&frames.Data{
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
