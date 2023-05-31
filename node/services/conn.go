package services

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/streams"
	"io"
)

type Conn struct {
	query    string
	link     *link.Link
	outbound bool
	remoteID id.Identity
	io.ReadWriteCloser
}

func (conn Conn) RemoteIdentity() id.Identity {
	return conn.remoteID
}

func (conn Conn) Query() string {
	return conn.query
}

func (conn Conn) Link() *link.Link {
	return conn.link
}

func (conn Conn) Outbound() bool {
	return conn.outbound
}

// pipe creates a pair of conns that talk to each other
func pipe(query string, link *link.Link) (*Conn, *Conn) {
	l, r := streams.Pipe()

	return &Conn{
			query:           query,
			link:            link,
			ReadWriteCloser: l,
			outbound:        true,
		}, &Conn{
			query:           query,
			link:            link,
			ReadWriteCloser: r,
		}
}
