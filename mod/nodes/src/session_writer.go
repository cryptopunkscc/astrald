package nodes

import (
	"errors"
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nodes/frames"
)

type sessionWriter struct {
	cond *sync.Cond

	paused bool

	buf    *OutputBuffer
	stream Stream
	nonce  astral.Nonce
}

func NewSessionWriter(buf *OutputBuffer, stream Stream, nonce astral.Nonce) *sessionWriter {
	w := &sessionWriter{
		buf:    buf,
		stream: stream,
		nonce:  nonce,
	}
	w.cond = sync.NewCond(&sync.Mutex{})
	return w
}

func (w *sessionWriter) Write(p []byte) (int, error) {
	total := 0

	for len(p) > 0 {
		w.cond.L.Lock()
		for w.paused {
			w.cond.Wait()
		}
		w.cond.L.Unlock()

		chunk, err := w.buf.Write(p, maxPayloadSize)
		if err != nil {
			var empty *ErrBufferEmpty
			if errors.As(err, &empty) {
				<-empty.ch
				continue
			}
			return total, err
		}

		if err := w.stream.Write(&frames.Data{
			Nonce:   w.nonce,
			Payload: chunk,
		}); err != nil {
			return total, err
		}

		n := len(chunk)
		total += n
		p = p[n:]
	}

	return total, nil
}

// Grow is called by mux on incoming window update (Read frame)
func (w *sessionWriter) Grow(n int) {
	w.buf.Grow(n)
}

func (w *sessionWriter) Close() {
	w.buf.Close()
}

// SetStream updates the stream and nonce under the cond lock. Must be called while paused.
func (w *sessionWriter) SetStream(stream Stream, nonce astral.Nonce) {
	w.cond.L.Lock()
	w.stream = stream
	w.nonce = nonce
	w.cond.L.Unlock()
}

func (w *sessionWriter) Pause() {
	w.cond.L.Lock()
	w.paused = true
	w.cond.L.Unlock()
}

func (w *sessionWriter) Resume() {
	w.cond.L.Lock()
	w.paused = false
	w.cond.Broadcast()
	w.cond.L.Unlock()
}
