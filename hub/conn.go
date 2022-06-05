package hub

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/link"
	"io"
)

type Conn struct {
	query string
	link  *link.Link
	io.ReadCloser
	io.WriteCloser
}

func (conn Conn) Query() string {
	return conn.query
}

func (conn Conn) Link() *link.Link {
	return conn.link
}

func (conn Conn) RemoteIdentity() id.Identity {
	if conn.link == nil {
		return id.Identity{}
	}
	return conn.link.RemoteIdentity()
}

func (conn Conn) IsLocal() bool {
	return conn.link == nil
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

// connPipe creates a pair of conns that talk to each other
func connPipe(query string, link *link.Link) (l Conn, r Conn) {
	// Set up a bidirectional stream using two pipes
	l.ReadCloser, r.WriteCloser = io.Pipe()
	r.ReadCloser, l.WriteCloser = io.Pipe()

	l.query, l.link = query, link
	r.query, r.link = query, link

	return
}
