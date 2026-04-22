package nodes

import (
	"errors"
	"io"
	"sync"
)

var _ io.WriteCloser = &sessionWriter{}

type sessionWriter struct {
	cond    *sync.Cond
	paused  bool
	writing bool
	buf     *OutputBuffer
}

func newSessionWriter(buf *OutputBuffer) *sessionWriter {
	w := &sessionWriter{buf: buf}
	w.cond = sync.NewCond(&sync.Mutex{})
	return w
}

func (w *sessionWriter) SetBuf(buf *OutputBuffer) {
	w.cond.L.Lock()
	w.buf = buf
	w.cond.L.Unlock()
}

func (w *sessionWriter) Close() error {
	w.cond.L.Lock()
	buf := w.buf
	w.cond.L.Unlock()
	buf.Close()

	return nil
}

func (w *sessionWriter) Grow(n int) {
	w.cond.L.Lock()
	buf := w.buf
	w.cond.L.Unlock()
	buf.Grow(n)
}

func (w *sessionWriter) Pause() {
	w.cond.L.Lock()
	w.paused = true
	for w.writing {
		w.cond.Wait()
	}
	w.cond.L.Unlock()
}

func (w *sessionWriter) Resume() {
	w.cond.L.Lock()
	w.paused = false
	w.cond.Broadcast()
	w.cond.L.Unlock()
}

func (w *sessionWriter) Write(p []byte) (int, error) {
	total := 0

	for len(p) > 0 {
		w.cond.L.Lock()
		for w.paused {
			w.cond.Wait()
		}
		buf := w.buf
		w.writing = true
		w.cond.L.Unlock()

		n, err := buf.Write(p)

		w.cond.L.Lock()
		w.writing = false
		w.cond.Broadcast()
		w.cond.L.Unlock()

		if err != nil {
			var empty *ErrBufferEmpty
			if errors.As(err, &empty) {
				<-empty.Wait()
				continue
			}
			return total, err
		}
		total += n
		p = p[n:]
	}

	return total, nil
}
