package mux

import (
	"errors"
	"io"
	"sync"
)

type InputStream struct {
	id      int
	r       io.Reader
	w       io.WriteCloser
	closeCh chan struct{}
	mu      sync.Mutex
	closed  bool
}

var _ io.Reader = &InputStream{}

func newInputStream(streamID int) *InputStream {
	r, w := io.Pipe()
	return &InputStream{
		id:      streamID,
		r:       r,
		w:       w,
		closeCh: make(chan struct{}),
	}
}

func (stream *InputStream) ID() int {
	return stream.id
}

func (stream *InputStream) Read(p []byte) (n int, err error) {
	return stream.r.Read(p)
}

func (stream *InputStream) Close() error {
	stream.mu.Lock()
	defer stream.mu.Unlock()

	if stream.closed {
		return errors.New("already closed")
	}

	defer close(stream.closeCh)
	stream.closed = true
	return stream.w.Close()
}

func (stream *InputStream) write(p []byte) (n int, err error) {
	return stream.w.Write(p)
}

// Wait returns a channel that will close when the InputStream closes
func (stream *InputStream) Wait() <-chan struct{} {
	return stream.closeCh
}
