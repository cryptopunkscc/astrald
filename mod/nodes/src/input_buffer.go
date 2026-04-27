package nodes

import (
	"io"
	"sync"

	"github.com/cryptopunkscc/astrald/mod/nodes"
)

var _ io.ReadCloser = &InputBuffer{}

// InputBuffer is a bounded receive buffer. Push appends data; Read consumes it.
// Non-blocking: returns ErrBufferEmpty when empty. onRead fires after each read
// to report consumed bytes. Closed buffer can still be read until drained.
//
// Lifecycle signals:
//   - Closed() fires when no more bytes will be pushed (Close called).
//   - Done()   fires when closed AND fully drained (no unread bytes remain).
type InputBuffer struct {
	mu      sync.Mutex
	size    int
	used    int
	buf     [][]byte
	ready   chan struct{}
	shut    chan struct{}
	done    chan struct{}
	drained bool
	closed  bool
	onRead  func(int)
}

func NewInputBuffer(size int, onRead func(int)) *InputBuffer {
	return &InputBuffer{size: size, onRead: onRead, shut: make(chan struct{}), done: make(chan struct{})}
}

func (b *InputBuffer) Push(p []byte) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nodes.ErrBufferClosed
	}
	if b.used+len(p) > b.size {
		return nodes.ErrBufferOverflow
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
	if b.closed {
		return nil
	}
	b.closed = true
	b.onRead = nil
	close(b.shut)
	b.signal()
	b.checkDone()

	return nil
}

func (b *InputBuffer) IsEmpty() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.used == 0
}

func (b *InputBuffer) Closed() <-chan struct{} {
	return b.shut
}

func (b *InputBuffer) Done() <-chan struct{} {
	return b.done
}

func (b *InputBuffer) checkDone() {
	if !b.drained && b.closed && len(b.buf) == 0 {
		b.drained = true
		close(b.done)
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
