package nodes

import (
	"errors"
	"io"
	"sync"
)

var _ io.ReadCloser = &sessionReader{}

type sessionReader struct {
	cond       *sync.Cond
	paused     bool
	buf        *InputBuffer
	nextBuffer *InputBuffer
}

func newSessionReader(buf *InputBuffer) *sessionReader {
	r := &sessionReader{buf: buf}
	r.cond = sync.NewCond(&sync.Mutex{})
	return r
}

func (r *sessionReader) SetNextBuffer(buf *InputBuffer) {
	r.cond.L.Lock()
	r.nextBuffer = buf
	r.cond.L.Unlock()
}

func (r *sessionReader) Close() error {
	r.cond.L.Lock()
	buf := r.buf
	r.cond.L.Unlock()
	buf.Close()

	return nil
}

func (r *sessionReader) Push(p []byte) error {
	r.cond.L.Lock()
	buf := r.buf
	r.cond.L.Unlock()
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
		buf := r.buf
		r.cond.L.Unlock()

		n, err = buf.Read(p)
		if err == nil {
			return
		}
		var empty *ErrBufferEmpty
		if !errors.As(err, &empty) {
			r.cond.L.Lock()
			if r.nextBuffer != nil {
				r.buf = r.nextBuffer
				r.nextBuffer = nil
				r.cond.L.Unlock()
				continue
			}
			r.cond.L.Unlock()
			return
		}

		<-empty.Wait()
	}
}
