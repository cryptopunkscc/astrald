package astral

import (
	"github.com/cryptopunkscc/astrald/id"
	"net"
)

var _ net.Conn = &Conn{}

type Conn struct {
	net.Conn
	remoteID id.Identity
	query    string
}

func (conn Conn) RemoteIdentity() id.Identity {
	return conn.remoteID
}

func (conn Conn) RemoteAddr() net.Addr {
	return Addr{address: conn.remoteID.String()}
}

func (conn Conn) Query() string {
	return conn.query
}
