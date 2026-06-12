package streams

import (
	"io"
)

var _ io.ReadWriteCloser = PipeEnd{}

// PipeEnd is one end of a bidirectional pipe; reading from it consumes data written by the remote end and vice versa.
// A read or write error on one side automatically closes the opposite direction.
type PipeEnd struct {
	r *io.PipeReader
	w *io.PipeWriter
}

// Pipe builds a bidirectional pipe and returns both ends of it
func Pipe() (left PipeEnd, right PipeEnd) {
	lr, rw := io.Pipe()
	rr, lw := io.Pipe()

	return PipeEnd{
			r: lr,
			w: lw,
		}, PipeEnd{
			r: rr,
			w: rw,
		}
}

// Read reads from the remote writer's data; closes the local write side on any error so the remote end sees EOF.
func (pipe PipeEnd) Read(p []byte) (n int, err error) {
	n, err = pipe.r.Read(p)
	if err != nil {
		pipe.w.Close()
	}
	return
}

// Write sends data to the remote reader; closes the local read side on any error.
func (pipe PipeEnd) Write(p []byte) (n int, err error) {
	n, err = pipe.w.Write(p)
	if err != nil {
		pipe.r.Close()
	}
	return
}

func (pipe PipeEnd) Close() error {
	pipe.r.Close()
	return pipe.w.Close()
}
