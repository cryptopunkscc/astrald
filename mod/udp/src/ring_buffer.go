package udp

import (
	"io"
	"sync"
)

// ringBuffer implements a blocking circular buffer for byte streams.
type ringBuffer struct {
	buf    []byte     // underlying buffer
	cap    int        // capacity (fixed)
	n      int        // current bytes stored
	r      int        // read position
	w      int        // write position
	mu     sync.Mutex // protects all fields
	notEmp *sync.Cond // signaled when buffer becomes non-empty
	notFul *sync.Cond // signaled when buffer has space available
	closed bool       // whether buffer is closed
}

// newRingBuffer creates a new ring buffer with the specified capacity.
func newRingBuffer(capacity int) *ringBuffer {
	if capacity < 0 {
		capacity = 0
	}
	rb := &ringBuffer{
		buf: make([]byte, capacity),
		cap: capacity,
	}
	rb.notEmp = sync.NewCond(&rb.mu)
	rb.notFul = sync.NewCond(&rb.mu)
	return rb
}

// WriteAll blocks until all bytes are written or the buffer is closed.
func (rb *ringBuffer) WriteAll(b []byte) (int, error) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	written := 0
	for written < len(b) {
		for rb.n == rb.cap && !rb.closed {
			rb.notFul.Wait()
		}
		if rb.closed {
			return written, io.ErrClosedPipe
		}

		space := rb.cap - rb.n
		toWrite := len(b) - written
		if toWrite > space {
			toWrite = space
		}

		end := (rb.w + toWrite) % rb.cap
		if end > rb.w {
			copy(rb.buf[rb.w:end], b[written:written+toWrite])
		} else {
			copy(rb.buf[rb.w:], b[written:written+toWrite])
			copy(rb.buf[:end], b[written+rb.cap-rb.w:written+toWrite])
		}

		rb.w = end
		rb.n += toWrite
		written += toWrite
		rb.notEmp.Signal()
	}

	return written, nil
}

// TryWrite attempts to write bytes without blocking.
func (rb *ringBuffer) TryWrite(b []byte) int {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if rb.n == rb.cap || rb.closed {
		return 0
	}

	space := rb.cap - rb.n
	toWrite := len(b)
	if toWrite > space {
		toWrite = space
	}

	end := (rb.w + toWrite) % rb.cap
	if end > rb.w {
		copy(rb.buf[rb.w:end], b[:toWrite])
	} else {
		copy(rb.buf[rb.w:], b[:toWrite])
		copy(rb.buf[:end], b[rb.cap-rb.w:toWrite])
	}

	rb.w = end
	rb.n += toWrite
	rb.notEmp.Signal()

	return toWrite
}

// Read blocks until at least one byte is available or the buffer is closed and drained.
func (rb *ringBuffer) Read(p []byte) (int, error) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	for rb.n == 0 && !rb.closed {
		rb.notEmp.Wait()
	}

	if rb.n == 0 && rb.closed {
		return 0, io.EOF
	}

	toRead := len(p)
	if toRead > rb.n {
		toRead = rb.n
	}

	end := (rb.r + toRead) % rb.cap
	if end > rb.r {
		copy(p, rb.buf[rb.r:end])
	} else {
		copy(p, rb.buf[rb.r:])
		copy(p[rb.cap-rb.r:], rb.buf[:end])
	}

	rb.r = end
	rb.n -= toRead
	rb.notFul.Signal()

	return toRead, nil
}

// Close marks the buffer as closed and wakes all waiters.
func (rb *ringBuffer) Close() {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if !rb.closed {
		rb.closed = true
		rb.notEmp.Broadcast()
		rb.notFul.Broadcast()
	}
}

// Len returns the number of bytes currently stored in the buffer.
func (rb *ringBuffer) Len() int {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	return rb.n
}

// Cap returns the capacity of the buffer.
func (rb *ringBuffer) Cap() int {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	return rb.cap
}
