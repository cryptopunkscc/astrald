package nodes

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nodes/frames"
)

func (mod *Peers) newSession(nonce astral.Nonce, remoteID *astral.Identity, queryStr string) (*session, bool) {
	sess, ok := mod.sessions.Set(nonce, newSession(nonce))
	if !ok {
		return sess, false
	}
	sess.RemoteIdentity = remoteID
	sess.Query = queryStr
	return sess, true
}

func (mod *Peers) newSessionBuffers(s *Stream, nonce astral.Nonce, sess *session) (*sessionReader, *sessionWriter) {
	inBuf, outBuf := mod.newBuffers(s, nonce, sess)
	return newSessionReader(inBuf), newSessionWriter(outBuf)
}

func (mod *Peers) newBuffers(s *Stream, nonce astral.Nonce, sess *session) (*InputBuffer, *OutputBuffer) {
	inBuf := NewInputBuffer(defaultBufferSize, func(n int) {
		sess.bytes.Add(uint64(n))
		s.Write(&frames.Read{Nonce: nonce, Len: uint32(n)})
	})

	outBuf := NewOutputBuffer(func(p []byte) error {
		remaining := p

		for len(remaining) > 0 {
			chunkSize := maxPayloadSize
			if len(remaining) < chunkSize {
				chunkSize = len(remaining)
			}

			chunk := remaining[:chunkSize]

			sess.bytes.Add(uint64(len(chunk)))

			if err := s.Write(&frames.Data{
				Nonce:   nonce,
				Payload: chunk,
			}); err != nil {
				return err
			}

			remaining = remaining[chunkSize:]
		}

		return nil
	})

	return inBuf, outBuf
}
