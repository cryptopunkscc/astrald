package mux

import (
	"io"
)

type InputStream struct {
	id      int
	r       io.Reader
	w       io.WriteCloser
	closeCh chan struct{}
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

func (stream *InputStream) StreamID() int {
	return stream.id
}

func (stream *InputStream) Read(p []byte) (n int, err error) {
	return stream.r.Read(p)
}

func (stream *InputStream) write(p []byte) (n int, err error) {
	return stream.w.Write(p)
}

func (stream *InputStream) Close() error {
	defer close(stream.closeCh) //FIXME: panic: close of closed channel

	return stream.w.Close()
}

func (stream *InputStream) WaitClose() <-chan struct{} {
	return stream.closeCh
}
