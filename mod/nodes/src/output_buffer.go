package nodes

import (
	"errors"
	"sync"
)

type OutputBuffer struct {
	mu     sync.Mutex
	wsize  int
	ready  chan struct{}
	closed bool
}

func NewOutputBuffer() *OutputBuffer {
	// fixme: add configurability
	return &OutputBuffer{}
}

func (b *OutputBuffer) Write(p []byte, maxChunk int) ([]byte, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil, errors.New("buffer closed")
	}
	if b.wsize == 0 {
		return nil, &ErrBufferEmpty{ch: b.readyCh()}
	}

	n := min(b.wsize, len(p), maxChunk)
	b.wsize -= n
	return p[:n], nil
}

func (b *OutputBuffer) Size() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.wsize
}

// Grow increases the remote window by n and wakes any waiting writers.
func (b *OutputBuffer) Grow(n int) {
	b.mu.Lock()
	b.wsize += n
	b.signal()
	b.mu.Unlock()
}

// Close marks the buffer closed and unblocks all waiters.
func (b *OutputBuffer) Close() {
	b.mu.Lock()
	b.closed = true
	b.signal()
	b.mu.Unlock()
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
