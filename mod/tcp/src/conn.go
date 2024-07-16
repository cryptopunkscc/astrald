package tcp

import (
	"github.com/cryptopunkscc/astrald/mod/exonet"
	_net "net"
)

var _ exonet.Conn = Conn{}

type Conn struct {
	_net.Conn
	outbound       bool
	localEndpoint  *Endpoint
	remoteEndpoint *Endpoint
}

// wrapTCPConn wraps a standard net.Conn into a astral's net.Conn with the addition of boundness
func wrapTCPConn(conn _net.Conn, outbound bool) *Conn {
	c := &Conn{
		Conn:     conn,
		outbound: outbound,
	}

	c.localEndpoint, _ = Parse(conn.LocalAddr().String())
	c.remoteEndpoint, _ = Parse(conn.RemoteAddr().String())

	return c
}

func (conn Conn) LocalEndpoint() exonet.Endpoint {
	return conn.localEndpoint
}

func (conn Conn) RemoteEndpoint() exonet.Endpoint {
	return conn.remoteEndpoint
}

func (conn Conn) Outbound() bool {
	return conn.outbound
}
