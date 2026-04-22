package nodes

import (
	"errors"
	"io"
	"sync"
)

var _ io.ReadCloser = &sessionReader{}

type sessionReader struct {
	cond   *sync.Cond
	paused bool
	mu     sync.Mutex
	buf    *InputBuffer

	nextBuffer *InputBuffer
}

func newSessionReader(buf *InputBuffer) *sessionReader {
	r := &sessionReader{buf: buf}
	r.cond = sync.NewCond(&sync.Mutex{})
	return r
}

func (r *sessionReader) SetBuf(buf *InputBuffer) {
	r.mu.Lock()
	r.buf = buf
	r.mu.Unlock()
}

func (r *sessionReader) SetNextBuffer(buf *InputBuffer) {
	r.mu.Lock()
	r.nextBuffer = buf
	r.mu.Unlock()
}

func (r *sessionReader) Close() error {
	r.mu.Lock()
	buf := r.buf
	r.mu.Unlock()
	buf.Close()

	return nil
}

func (r *sessionReader) Push(p []byte) error {
	r.mu.Lock()
	buf := r.buf
	r.mu.Unlock()
	return buf.Push(p)
}

func (r *sessionReader) Pause() {
	r.cond.L.Lock()
	defer r.cond.L.Unlock()
	r.paused = true
}

func (r *sessionReader) Resume() {
	r.cond.L.Lock()
	defer r.cond.L.Unlock()
	r.paused = false
	r.cond.Broadcast()
}

func (r *sessionReader) Read(p []byte) (n int, err error) {
	for {
		r.cond.L.Lock()
		for r.paused {
			r.cond.Wait()
		}
		r.cond.L.Unlock()

		r.mu.Lock()
		buf := r.buf
		r.mu.Unlock()

		n, err = buf.Read(p)
		if err == nil {
			return
		}
		var empty *ErrBufferEmpty
		if !errors.As(err, &empty) {
			r.mu.Lock()
			if r.nextBuffer != nil {
				r.buf = r.nextBuffer
				r.nextBuffer = nil
				r.mu.Unlock()
				continue
			}
			r.mu.Unlock()
			return
		}
		<-empty.Wait()
	}
}
