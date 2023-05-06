package inet

import (
	"github.com/cryptopunkscc/astrald/net"
	_net "net"
)

var _ net.Conn = Conn{}

type Conn struct {
	_net.Conn
	outbound bool
}

// newConn wraps a standard net.Conn into a astral's net.Conn with the addition of boundness
func newConn(conn _net.Conn, outbound bool) Conn {
	return Conn{
		Conn:     conn,
		outbound: outbound,
	}
}

func (conn Conn) LocalEndpoint() net.Endpoint {
	e, _ := Parse(conn.Conn.LocalAddr().String())
	return e
}

func (conn Conn) RemoteEndpoint() net.Endpoint {
	e, _ := Parse(conn.Conn.RemoteAddr().String())
	return e
}

func (conn Conn) Outbound() bool {
	return conn.outbound
}
