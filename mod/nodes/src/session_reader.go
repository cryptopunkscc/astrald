package nodes

import (
	"errors"
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nodes/frames"
)

type sessionReader struct {
	mu     sync.Mutex
	buf    *InputBuffer
	stream *Stream
	nonce  astral.Nonce
}

func newSessionReader(buf *InputBuffer, stream *Stream, nonce astral.Nonce) *sessionReader {
	return &sessionReader{buf: buf, stream: stream, nonce: nonce}
}

func (r *sessionReader) SetStream(stream *Stream, nonce astral.Nonce) {
	r.mu.Lock()
	r.stream = stream
	r.nonce = nonce
	r.mu.Unlock()
}

func (r *sessionReader) Close() {
	r.buf.Close()
}

func (r *sessionReader) Read(p []byte) (n int, err error) {
	for {
		n, err = r.buf.Read(p)
		if err == nil {
			r.mu.Lock()
			stream, nonce := r.stream, r.nonce
			r.mu.Unlock()
			stream.Write(&frames.Read{Nonce: nonce, Len: uint32(n)})
			return
		}
		var empty *ErrBufferEmpty
		if !errors.As(err, &empty) {
			return
		}
		<-empty.Wait()
	}
}
