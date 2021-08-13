package mux

import (
	"sync"
)

type OutputStream struct {
	id      int
	mux     *Mux
	mu      sync.Mutex
	closeCh chan struct{}
}

func NewOutputStream(mux *Mux, remoteStreamID int) *OutputStream {
	return &OutputStream{
		id:      remoteStreamID,
		mux:     mux,
		closeCh: make(chan struct{}),
	}
}

// Write writes a byte buffer to the connectiontion
func (stream *OutputStream) Write(data []byte) (n int, err error) {
	stream.mu.Lock()
	defer stream.mu.Unlock()

	if stream.mux == nil {
		return 0, ErrStreamClosed
	}

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

func (stream *OutputStream) Close() (err error) {
	stream.mu.Lock()
	defer stream.mu.Unlock()

	err = stream.mux.Write(stream.id, nil)
	stream.mux = nil

	close(stream.closeCh)
	return
}

func (stream OutputStream) StreamID() int {
	return stream.id
}

func (stream *OutputStream) WaitClose() <-chan struct{} {
	return stream.closeCh
}
