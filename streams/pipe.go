package streams

import (
	"errors"
	"io"
)

var _ io.ReadWriteCloser = pipeEnd{}

type pipeEnd struct {
	io.ReadCloser
	io.WriteCloser
}

// Pipe builds a bidirectional pipe and returns both ends of it
func Pipe() (left io.ReadWriteCloser, right io.ReadWriteCloser) {
	lr, rw := io.Pipe()
	rr, lw := io.Pipe()

	return pipeEnd{
			ReadCloser:  lr,
			WriteCloser: lw,
		}, pipeEnd{
			ReadCloser:  rr,
			WriteCloser: rw,
		}
}

// io.ErrClosedPipe, io.EOF
func (pipe pipeEnd) Read(p []byte) (n int, err error) {
	n, err = pipe.ReadCloser.Read(p)
	switch {
	case err == nil:
	case errors.Is(err, io.EOF): // EOF is not a fatal error
	default:
		pipe.WriteCloser.Close()
	}
	return
}

func (pipe pipeEnd) Write(p []byte) (n int, err error) {
	n, err = pipe.WriteCloser.Write(p)
	switch {
	case err == nil:
	default:
		pipe.ReadCloser.Close()
	}
	return
}

func (pipe pipeEnd) Close() error {
	return pipe.WriteCloser.Close()
}
