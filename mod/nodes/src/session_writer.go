package nodes

import (
	"errors"
	"sync"
)

// sessionWriter is a blocking io.Writer for a session.
// It owns only the pause gate and blocking loop; frame writing is handled by OutputBuffer.write.
type sessionWriter struct {
	cond   *sync.Cond
	paused bool
	buf    *OutputBuffer
}

func NewSessionWriter(buf *OutputBuffer) *sessionWriter {
	w := &sessionWriter{buf: buf}
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

		n, err := w.buf.Write(p, maxPayloadSize)
		if err != nil {
			var empty *ErrBufferEmpty
			if errors.As(err, &empty) {
				<-empty.ch
				continue
			}
			return total, err
		}
		total += n
		p = p[n:]
	}

	return total, nil
}

func (w *sessionWriter) Close() {
	w.buf.Close()
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
