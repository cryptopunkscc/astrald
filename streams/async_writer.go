package streams

import (
	"errors"
	"io"
	"sync"
)

var ErrBufferOverflow = errors.New("buffer overflow")

type AsyncWriter struct {
	writer     io.WriteCloser
	afterFlush func([]byte)
	cond       *sync.Cond

	chunks     [][]byte
	bufferSize int
	used       int
	writeErr   error
	closed     bool
	done       chan struct{}
}

func NewAsyncWriter(output io.WriteCloser, bufferSize int) *AsyncWriter {
	var w = &AsyncWriter{
		writer:     output,
		chunks:     make([][]byte, 0),
		bufferSize: bufferSize,
		done:       make(chan struct{}),
		cond:       sync.NewCond(&sync.Mutex{}),
	}

	go w.flusher()

	return w

}

func (b *AsyncWriter) Write(p []byte) (int, error) {
	b.cond.L.Lock()
	defer b.cond.L.Unlock()

	if b.writeErr != nil {
		return 0, b.writeErr
	}

	if b.closed {
		return 0, io.ErrClosedPipe
	}

	if len(p) > b.bufferSize-b.used {
		return 0, ErrBufferOverflow
	}

	b.chunks = append(b.chunks, p)
	b.used += len(p)

	b.cond.Broadcast()

	return len(p), nil
}

func (b *AsyncWriter) Close() error {
	b.cond.L.Lock()
	defer b.cond.L.Unlock()

	b.closed = true

	b.cond.Broadcast()

	return nil
}

func (b *AsyncWriter) Used() int {
	b.cond.L.Lock()
	defer b.cond.L.Unlock()

	return b.used
}

func (b *AsyncWriter) AfterFlush() func([]byte) {
	b.cond.L.Lock()
	defer b.cond.L.Unlock()

	return b.afterFlush
}

func (b *AsyncWriter) SetAfterFlush(afterFlush func([]byte)) {
	b.cond.L.Lock()
	defer b.cond.L.Unlock()

	b.afterFlush = afterFlush
}

func (b *AsyncWriter) BufferSize() int {
	b.cond.L.Lock()
	defer b.cond.L.Unlock()

	return b.bufferSize
}

func (b *AsyncWriter) SetBufferSize(bufferSize int) {
	b.cond.L.Lock()
	defer b.cond.L.Unlock()

	b.bufferSize = bufferSize
}

func (b *AsyncWriter) Writer() io.WriteCloser {
	b.cond.L.Lock()
	defer b.cond.L.Unlock()

	return b.writer
}

func (b *AsyncWriter) SetWriter(output io.WriteCloser) {
	b.cond.L.Lock()
	defer b.cond.L.Unlock()

	b.writer = output
}

func (b *AsyncWriter) Sync() error {
	b.cond.L.Lock()
	defer b.cond.L.Unlock()

	for {
		if b.writeErr != nil {
			return b.writeErr
		}
		if len(b.chunks) == 0 {
			return nil
		}
		b.cond.Wait()
	}
}

func (b *AsyncWriter) Done() <-chan struct{} {
	return b.done
}

func (b *AsyncWriter) Err() error {
	return b.writeErr
}

func (b *AsyncWriter) flusher() {
	defer func() {
		b.Writer().Close()
		close(b.done)
		b.cond.Broadcast()
	}()

	b.cond.L.Lock()
	for {
		for len(b.chunks) > 0 {
			chunk := b.chunks[0]
			b.cond.L.Unlock()

			for len(chunk) > 0 {
				n, err := b.Writer().Write(chunk)

				if err != nil {
					b.writeErr = err
					return
				}
				chunk = chunk[n:]
			}

			chunk = b.dequeue()
			if b.afterFlush != nil {
				b.afterFlush(chunk)
			}
			b.cond.L.Lock()
		}
		if b.closed {
			b.cond.L.Unlock()
			return
		}

		// wait for any change in the buffer
		b.cond.Wait()
	}
}

func (b *AsyncWriter) dequeue() []byte {
	b.cond.L.Lock()
	defer b.cond.L.Unlock()

	if len(b.chunks) == 0 {
		return nil
	}

	var chunk = b.chunks[0]
	b.chunks = b.chunks[1:]
	b.used -= len(chunk)

	b.cond.Broadcast()

	return chunk
}
