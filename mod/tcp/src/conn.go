package tcp

import (
	"net"

	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/tcp"
)

var _ exonet.Conn = Conn{}

type Conn struct {
	net.Conn
	outbound       bool
	localEndpoint  *tcp.Endpoint
	remoteEndpoint *tcp.Endpoint
}

// wrapTCPConn wraps a standard net.Conn into a astral's net.Conn with the addition of boundness
func wrapTCPConn(conn net.Conn, outbound bool) *Conn {
	c := &Conn{
		Conn:     conn,
		outbound: outbound,
	}

	c.localEndpoint, _ = tcp.ParseEndpoint(conn.LocalAddr().String())
	c.remoteEndpoint, _ = tcp.ParseEndpoint(conn.RemoteAddr().String())

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
