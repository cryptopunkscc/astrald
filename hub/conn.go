package hub

import (
	"io"
)

type Conn struct {
	io.ReadCloser
	io.WriteCloser
}

func (conn Conn) Read(p []byte) (n int, err error) {
	n, err = conn.ReadCloser.Read(p)
	if err != nil {
		_ = conn.WriteCloser.Close()
	}

	return
}

func (conn Conn) Close() error {
	return conn.WriteCloser.Close()
}

// pipe creates a pair of conns that talk to each other
func pipe() (left Conn, right Conn) {
	// Set up a bidirectional stream using two pipes
	left.ReadCloser, right.WriteCloser = io.Pipe()
	right.ReadCloser, left.WriteCloser = io.Pipe()

	return
}
