package streams

import (
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
	if err != nil {
		pipe.WriteCloser.Close()
	}
	return
}

func (pipe pipeEnd) Write(p []byte) (n int, err error) {
	n, err = pipe.WriteCloser.Write(p)
	if err != nil {
		pipe.ReadCloser.Close()
	}
	return
}

func (pipe pipeEnd) Close() error {
	pipe.ReadCloser.Close()
	return pipe.WriteCloser.Close()
}
