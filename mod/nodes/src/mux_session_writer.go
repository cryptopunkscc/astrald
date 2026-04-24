package nodes

import (
	"errors"
	"io"
	"sync"
)

var _ io.WriteCloser = &muxSessionWriter{}

type muxSessionWriter struct {
	cond   *sync.Cond
	closed bool
	paused bool
	buf    *OutputBuffer
}

func newSessionWriter(buf *OutputBuffer) *muxSessionWriter {
	w := &muxSessionWriter{buf: buf}
	w.cond = sync.NewCond(&sync.Mutex{})
	return w
}

func (w *muxSessionWriter) Buf() *OutputBuffer {
	w.cond.L.Lock()
	defer w.cond.L.Unlock()
	return w.buf
}

func (w *muxSessionWriter) SetBuf(buf *OutputBuffer) {
	w.cond.L.Lock()
	defer w.cond.L.Unlock()
	w.buf = buf
}

func (w *muxSessionWriter) Close() error {
	w.cond.L.Lock()
	if w.closed {
		w.cond.L.Unlock()
		return nil
	}
	w.closed = true
	buf := w.buf
	w.cond.Broadcast()
	w.cond.L.Unlock()
	buf.Close()

	return nil
}

func (w *muxSessionWriter) Grow(n int) {
	w.cond.L.Lock()
	buf := w.buf
	w.cond.L.Unlock()
	buf.Grow(n)
}

func (w *muxSessionWriter) Pause() {
	w.cond.L.Lock()
	defer w.cond.L.Unlock()
	w.paused = true
}

func (w *muxSessionWriter) Resume() {
	w.cond.L.Lock()
	defer w.cond.L.Unlock()
	w.paused = false
	w.cond.Broadcast()
}

func (w *muxSessionWriter) Write(p []byte) (int, error) {
	total := 0

	for len(p) > 0 {
		w.cond.L.Lock()

		for w.paused && !w.closed {
			w.cond.Wait()
		}
		if w.closed {
			w.cond.L.Unlock()
			return total, io.ErrClosedPipe
		}

		n, err := w.buf.Write(p)
		if err != nil {
			var emptyErr *ErrBufferEmpty
			if errors.As(err, &emptyErr) {
				w.cond.L.Unlock()
				<-emptyErr.Wait()
				continue
			}

			w.cond.L.Unlock()
			return total, err
		}

		total += n
		p = p[n:]

		w.cond.L.Unlock()
	}

	return total, nil
}
