package cslq

import (
	"io"
)

// ByteStackReader is a byte stack that reads a byte from the io.Reader when Pop() is called on an empty stack.
// Pop and Push are not thread safe.
type ByteStackReader struct {
	r   io.Reader
	buf []byte
}

// NewByteStackReader returns a new ByteStackReader with the provided io.Reader as the secondary source of bytes.
func NewByteStackReader(r io.Reader) *ByteStackReader {
	return &ByteStackReader{
		r:   r,
		buf: make([]byte, 0),
	}
}

// Pop removes and returns the top byte from the stack. If the stack is empty, Pop reads a byte from the io.Reader and
// returns it.
func (r *ByteStackReader) Pop() (b byte, err error) {
	if len(r.buf) > 0 {
		b = r.buf[len(r.buf)-1]
		r.buf = r.buf[:len(r.buf)-1]
		return
	}

	var buf [1]byte

	_, err = r.r.Read(buf[0:1])
	b = buf[0]
	return
}

// Push puts a byte on the top of the stack.
func (r *ByteStackReader) Push(b byte) {
	r.buf = append(r.buf, b)
}
