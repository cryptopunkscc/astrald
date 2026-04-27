package nodes

import (
	"errors"
	"io"
	"sync"
)

var _ io.ReadCloser = &InputBuffer{}

type InputBuffer struct {
	mu     sync.Mutex
	size   int
	used   int
	buf    [][]byte
	ready  chan struct{}
	done   chan struct{}
	closed bool
	onRead func(int)
}

func NewInputBuffer(size int, onRead func(int)) *InputBuffer {
	return &InputBuffer{size: size, onRead: onRead}
}

func (b *InputBuffer) Push(p []byte) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return errors.New("buffer closed")
	}
	if b.used+len(p) > b.size {
		return errors.New("buffer overflow")
	}

	b.buf = append(b.buf, p)
	b.used += len(p)
	b.signal()
	return nil
}

func (b *InputBuffer) Read(p []byte) (n int, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.buf) == 0 {
		if b.closed {
			return 0, io.EOF
		}

		return 0, &ErrBufferEmpty{ch: b.readyCh()}
	}

	chunk := b.buf[0]
	n = min(len(p), len(chunk))
	copy(p, chunk[:n])
	chunk = chunk[n:]
	b.used -= n
	if len(chunk) > 0 {
		b.buf[0] = chunk
	} else {
		b.buf = b.buf[1:]
	}

	if b.onRead != nil {
		b.onRead(n)
	}
	b.checkDone()
	return
}

func (b *InputBuffer) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.closed = true
	b.onRead = nil
	b.signal()
	b.checkDone()

	return nil
}

func (b *InputBuffer) IsEmpty() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.used == 0
}

func (b *InputBuffer) Done() <-chan struct{} {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.done == nil {
		b.done = make(chan struct{})
	}
	b.checkDone()
	return b.done
}

func (b *InputBuffer) checkDone() {
	if b.closed && len(b.buf) == 0 && b.done != nil {
		select {
		case <-b.done:
		default:
			close(b.done)
		}
	}
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
