package nodes

import (
	"errors"
	"sync"
)

// InputBuffer is a receive-side buffer for a session.
// Non-blocking: Read returns ErrBufferEmpty immediately when no data is available.
// onRead is called after each successful read (while the lock is held) so the
// caller can inject acknowledgement logic without the buffer knowing about frames.
type InputBuffer struct {
	mu     sync.Mutex
	rsize  int
	rused  int
	rbuf   [][]byte
	ready  chan struct{}
	closed bool
	onRead func(int)
}

func NewInputBuffer(size int, onRead func(int)) *InputBuffer {
	return &InputBuffer{rsize: size, onRead: onRead}
}

// Push appends a payload chunk to the buffer.
func (b *InputBuffer) Push(p []byte) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return errors.New("buffer closed")
	}
	if b.rused+len(p) > b.rsize {
		return errors.New("buffer overflow")
	}

	b.rbuf = append(b.rbuf, p)
	b.rused += len(p)
	b.signal()
	return nil
}

func (b *InputBuffer) Read(p []byte) (n int, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.rbuf) == 0 {
		if b.closed {
			return 0, errors.New("connection closed")
		}
		return 0, &ErrBufferEmpty{ch: b.readyCh()}
	}

	chunk := b.rbuf[0]
	n = min(len(p), len(chunk))
	copy(p, chunk[:n])
	chunk = chunk[n:]
	b.rused -= n
	if len(chunk) > 0 {
		b.rbuf[0] = chunk
	} else {
		b.rbuf = b.rbuf[1:]
	}

	if b.onRead != nil {
		b.onRead(n)
	}

	return
}

func (b *InputBuffer) Wait() <-chan struct{} {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.readyCh()
}

func (b *InputBuffer) Close() {
	b.mu.Lock()
	b.closed = true
	b.signal()
	b.mu.Unlock()
}

func (b *InputBuffer) readyCh() chan struct{} {
	if b.ready == nil {
		b.ready = make(chan struct{})
	}
	return b.ready
}

func (b *InputBuffer) signal() {
	if b.ready != nil {
		close(b.ready)
		b.ready = nil
	}
}
