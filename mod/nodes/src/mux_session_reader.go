package nodes

import (
	"errors"
	"io"
	"sync"
)

var _ io.ReadCloser = &muxSessionReader{}

type muxSessionReader struct {
	cond       *sync.Cond
	paused     bool
	buf        *InputBuffer
	nextBuffer *InputBuffer
}

func newSessionReader(buf *InputBuffer) *muxSessionReader {
	r := &muxSessionReader{buf: buf}
	r.cond = sync.NewCond(&sync.Mutex{})
	return r
}

func (r *muxSessionReader) Buf() *InputBuffer {
	r.cond.L.Lock()
	defer r.cond.L.Unlock()
	return r.buf
}

func (r *muxSessionReader) Close() error {
	r.cond.L.Lock()
	buf := r.buf
	nextBuffer := r.nextBuffer
	r.nextBuffer = nil
	r.paused = false
	r.cond.Broadcast()
	r.cond.L.Unlock()

	if buf != nil {
		_ = buf.Close()
	}
	if nextBuffer != nil {
		_ = nextBuffer.Close()
	}

	return nil
}

func (r *muxSessionReader) CloseBuf() error {
	r.cond.L.Lock()
	buf := r.buf
	r.cond.L.Unlock()
	return buf.Close()
}

func (r *muxSessionReader) SetNextBuffer(buf *InputBuffer) {
	r.cond.L.Lock()
	defer r.cond.L.Unlock()
	r.nextBuffer = buf
}

func (r *muxSessionReader) Pause() {
	r.cond.L.Lock()
	defer r.cond.L.Unlock()
	r.paused = true
	r.cond.Broadcast()
}

func (r *muxSessionReader) Resume() {
	r.cond.L.Lock()
	defer r.cond.L.Unlock()
	r.paused = false
	r.cond.Broadcast()
}

func (r *muxSessionReader) Read(p []byte) (n int, err error) {
	for {
		r.cond.L.Lock()
		for r.paused {
			r.cond.Wait()
		}

		n, err := r.buf.Read(p)
		switch {
		case err == nil:
			r.cond.L.Unlock()
			return n, nil
		case err == io.EOF:
			if r.nextBuffer != nil {
				r.buf = r.nextBuffer
				r.nextBuffer = nil
				r.cond.L.Unlock()
				continue
			}

			r.cond.L.Unlock()
			return n, err
		default:
			var emptyErr *ErrBufferEmpty
			if errors.As(err, &emptyErr) {
				go func() {
					<-emptyErr.Wait()
					r.cond.Broadcast()
				}()
				r.cond.Wait()
				r.cond.L.Unlock()
				continue
			}

			r.cond.L.Unlock()
			return n, err
		}
	}
}
