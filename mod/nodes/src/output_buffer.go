package nodes

import (
	"errors"
	"io"
	"sync"
)

var _ io.WriteCloser = &OutputBuffer{}

type OutputBuffer struct {
	mu     sync.Mutex
	wsize  int
	ready  chan struct{}
	closed bool

	write func([]byte) error
}

func NewOutputBuffer(write func([]byte) error) *OutputBuffer {
	return &OutputBuffer{write: write}
}

func (b *OutputBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return 0, errors.New("buffer closed")
	}
	if b.wsize == 0 {
		return 0, &ErrBufferEmpty{ch: b.readyCh()}
	}

	n := min(b.wsize, len(p))
	b.wsize -= n

	return n, b.write(p[:n])
}

func (b *OutputBuffer) Size() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.wsize
}

func (b *OutputBuffer) Grow(n int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.wsize += n
	b.signal()
}

func (b *OutputBuffer) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.closed = true
	b.signal()

	return nil
}

func (b *OutputBuffer) readyCh() chan struct{} {
	if b.ready == nil {
		b.ready = make(chan struct{})
	}
	return b.ready
}

func (b *OutputBuffer) signal() {
	if b.ready != nil {
		close(b.ready)
		b.ready = nil
	}
}
