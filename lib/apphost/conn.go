package apphost

import (
	"net"

	"github.com/cryptopunkscc/astrald/astral"
)

type Conn struct {
	net.Conn
	query    *astral.Query
	outbound bool
}

var _ net.Conn = &Conn{}
var _ astral.Conn = &Conn{}

func NewConn(conn net.Conn, query *astral.Query, outbound bool) *Conn {
	return &Conn{Conn: conn, query: query, outbound: outbound}
}

func (conn Conn) Query() *astral.Query {
	return conn.query
}

func (conn Conn) RemoteIdentity() *astral.Identity {
	if conn.outbound {
		return conn.query.Target
	}
	return conn.query.Caller
}

func (conn Conn) LocalIdentity() *astral.Identity {
	if conn.outbound {
		return conn.query.Caller
	}
	return conn.query.Target
}

func (conn Conn) RemoteAddr() net.Addr {
	return Addr{address: conn.RemoteIdentity().String()}
}
