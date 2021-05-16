package hub

import (
	"io"
)

type Conn struct {
	io.ReadCloser
	io.WriteCloser
}

// pipe creates a pair of conns that talk to each other
func pipe() (a Conn, b Conn) {
	// Set up a bidirectional stream using two pipes
	a.ReadCloser, b.WriteCloser = io.Pipe()
	b.ReadCloser, a.WriteCloser = io.Pipe()

	return
}

func (conn Conn) Close() error {
	_ = conn.ReadCloser.Close()
	_ = conn.WriteCloser.Close()
	return nil
}

func (conn Conn) Read(p []byte) (n int, err error) {
	n, err = conn.ReadCloser.Read(p)
	if err != nil {
		_ = conn.Close()
	}
	return
}
