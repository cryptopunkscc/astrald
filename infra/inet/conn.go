package inet

import (
	"github.com/cryptopunkscc/astrald/infra"
	"net"
)

var _ infra.Conn = Conn{}

type Conn struct {
	net.Conn
	outbound   bool
	remoteAddr Addr
}

// newConn wraps a standard net.Conn into a astral's net.Conn with the addition of boundness
func newConn(conn net.Conn, outbound bool) Conn {
	return Conn{
		Conn:     conn,
		outbound: outbound,
	}
}

func (conn Conn) LocalAddr() infra.Addr {
	addr, _ := Parse(conn.Conn.LocalAddr().String())
	return addr
}

func (conn Conn) RemoteAddr() infra.Addr {
	addr, _ := Parse(conn.Conn.RemoteAddr().String())
	return addr
}

func (conn Conn) Outbound() bool {
	return conn.outbound
}
