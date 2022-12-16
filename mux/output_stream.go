package mux

import (
	"sync"
)

type OutputStream struct {
	id     int
	mux    *Mux
	closed bool
	mu     sync.Mutex
	err    error
}

func NewOutputStream(mux *Mux, streamID int) *OutputStream {
	return &OutputStream{
		id:  streamID,
		mux: mux,
	}
}

// Write writes a byte buffer to the connectiontion
func (stream *OutputStream) Write(data []byte) (n int, err error) {
	stream.mu.Lock()
	defer stream.mu.Unlock()

	if stream.closed {
		return 0, ErrStreamClosed
	}

	return stream.write(data)
}

func (stream *OutputStream) write(data []byte) (n int, err error) {
	left := data[:]

	for len(left) > 0 {
		chunkLen := MaxPayload
		if chunkLen > len(left) {
			chunkLen = len(left)
		}

		err = stream.mux.Write(stream.id, left[0:chunkLen])
		if err != nil {
			return
		}

		n += chunkLen
		left = left[chunkLen:]
	}

	return
}

// Close closes the stream
func (stream *OutputStream) Close() (err error) {
	stream.mu.Lock()
	defer stream.mu.Unlock()

	if stream.closed {
		return nil
	}

	stream.mux.Write(stream.id, nil)
	stream.closed = true
	return
}

func (stream *OutputStream) ID() int {
	return stream.id
}
