package streams

import (
	"context"
	"io"
	"sync"
)

const ReadBufferSize = 16 * 1024

// ContextReader is a buffered cancelable reader. You cannot abort a Read() in Go, but you can stop waiting for the
// Read() to finish and buffer the data after it does.
type ContextReader struct {
	r   io.Reader
	buf []byte
	ch  chan struct{}
	mu  sync.Mutex
	err error
}

// NewContextReader returns a new instance of a ContextReader
func NewContextReader(r io.Reader) *ContextReader {
	return &ContextReader{r: r}
}

// ReadContext performs a Read on the underlying reader. If the context finishes before the read, the read will continue
// in the background and buffer the data once it's done. The next call to Read will read from the buffer
func (r *ContextReader) ReadContext(ctx context.Context, p []byte) (n int, err error) {
	for {
		r.mu.Lock()

		// read from the buffer first
		if len(r.buf) > 0 {
			n = min(len(p), len(r.buf))
			copy(p, r.buf[:n])
			r.buf = r.buf[n:]
			defer r.mu.Unlock()
			return n, nil
		}

		// once buffer is empty, check for errors
		if r.err != nil {
			defer r.mu.Unlock()
			return 0, r.err
		}

		// make sure a reader is running
		if r.ch == nil {
			r.ch = make(chan struct{})
			go func() {
				defer close(r.ch)

				var b = make([]byte, ReadBufferSize)
				n, err := r.r.Read(b)

				r.mu.Lock()
				defer r.mu.Unlock()

				r.err = err
				if n > 0 {
					r.buf = append(r.buf, b[:n]...)
				}
				r.ch = nil
			}()
		}

		r.mu.Unlock()

		select {
		case <-r.ch:
			continue

		case <-ctx.Done():
			return 0, ctx.Err()
		}
	}
}

// WithContext wraps the ContextReader and the Context in a wrapper that implements the regular io.Reader interface.
func (r *ContextReader) WithContext(ctx context.Context) io.Reader {
	return &readerContextWrapper{
		ctx: ctx,
		r:   r,
	}
}

type readerContextWrapper struct {
	ctx context.Context
	r   *ContextReader
}

func (r readerContextWrapper) Read(p []byte) (n int, err error) {
	return r.r.ReadContext(r.ctx, p)
}
