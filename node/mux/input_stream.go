package mux

import (
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

func newInputStream(streamID int) *InputStream {
	r, w := io.Pipe()
	return &InputStream{
		id:      streamID,
		r:       r,
		w:       w,
		closeCh: make(chan struct{}),
	}
}

func (stream *InputStream) StreamID() int {
	return stream.id
}

func (stream *InputStream) Read(p []byte) (n int, err error) {
	if stream.closed {
		return n, io.EOF
	}

	n, err = stream.r.Read(p)

	if err != nil {
		stream.Close()
	}
	return n, err
}

func (stream *InputStream) write(p []byte) (n int, err error) {
	return stream.w.Write(p)
}

func (stream *InputStream) Close() error {
	stream.mu.Lock()
	defer stream.mu.Unlock()

	if stream.closed {
		return nil
	}

	stream.closed = true
	stream.w.Close()
	close(stream.closeCh)

	return nil
}

func (stream *InputStream) WaitClose() <-chan struct{} {
	return stream.closeCh
}
