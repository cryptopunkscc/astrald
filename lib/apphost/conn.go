package apphost

import (
	"github.com/cryptopunkscc/astrald/astral"
	"net"
)

var _ net.Conn = &Conn{}

type Conn struct {
	net.Conn
	remoteID *astral.Identity
	query    string
}

func (conn Conn) RemoteIdentity() *astral.Identity {
	return conn.remoteID
}

func (conn Conn) RemoteAddr() net.Addr {
	return Addr{address: conn.remoteID.String()}
}

func (conn Conn) Query() string {
	return conn.query
}
