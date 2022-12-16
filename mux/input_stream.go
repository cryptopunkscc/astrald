package mux

import (
	"io"
	"sync"
	"sync/atomic"
)

type InputStream struct {
	id      int
	r       *io.PipeReader
	w       *io.PipeWriter
	closeCh chan struct{}
	closed  atomic.Bool
	mu      sync.Mutex
	demux   *StreamDemux
}

var _ io.Reader = &InputStream{}

func newInputStream(demux *StreamDemux, streamID int) *InputStream {
	r, w := io.Pipe()
	return &InputStream{
		demux:   demux,
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
	if stream.closed.Load() {
		return 0, ErrStreamClosed
	}

	return stream.r.Read(p)
}

// Discard discards an input stream without waiting for it to be closed in the multiplexed stream. Use this only if
// you allocated an input stream, but never actually used it.
func (stream *InputStream) Discard() (err error) {
	if stream.closed.CompareAndSwap(false, true) {
		close(stream.closeCh)
		stream.r.Close()
		return stream.demux.removeInputStream(stream.id)
	}
	return nil
}

// Wait returns a channel that will close when the InputStream closes
func (stream *InputStream) Wait() <-chan struct{} {
	return stream.closeCh
}

// closeWriter closes the pipe from the writer's (remote peer's) side
func (stream *InputStream) closeWriter(err error) error {
	if stream.closed.CompareAndSwap(false, true) {
		close(stream.closeCh)
		return stream.w.CloseWithError(err)
	}
	return nil
}

func (stream *InputStream) write(p []byte) (n int, err error) {
	if stream.closed.Load() {
		return 0, ErrStreamClosed
	}

	return stream.w.Write(p)
}
