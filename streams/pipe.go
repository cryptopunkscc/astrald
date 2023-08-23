package streams

import (
	"io"
)

var _ io.ReadWriteCloser = PipeEnd{}

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

// io.ErrClosedPipe, io.EOF
func (pipe PipeEnd) Read(p []byte) (n int, err error) {
	n, err = pipe.r.Read(p)
	if err != nil {
		pipe.w.Close()
	}
	return
}

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
