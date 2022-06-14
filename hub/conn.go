package hub

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/streams"
	"io"
)

type Conn struct {
	query string
	link  *link.Link
	io.ReadWriteCloser
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

// pipe creates a pair of conns that talk to each other
func pipe(query string, link *link.Link) (Conn, Conn) {
	l, r := streams.Pipe()

	return Conn{
			query:           query,
			link:            link,
			ReadWriteCloser: l,
		}, Conn{
			query:           query,
			link:            link,
			ReadWriteCloser: r,
		}
}
