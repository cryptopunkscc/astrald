package tor

import (
	"github.com/cryptopunkscc/astrald/infra"
	"net"
)

type Conn struct {
	net.Conn
	addr     Addr
	outbound bool
}

func (conn Conn) LocalAddr() infra.Addr {
	return Addr{}
}

var _ infra.Conn = Conn{}

// newConn wraps a standard net.Conn into a astral's net.Conn with the addition of boundness
func newConn(conn net.Conn, addr Addr, outbound bool) Conn {
	return Conn{
		Conn:     conn,
		addr:     addr,
		outbound: outbound,
	}
}

func (conn Conn) RemoteAddr() infra.Addr {
	return conn.addr
}

func (conn Conn) Outbound() bool {
	return conn.outbound
}
