package inet4

import (
	net2 "github.com/cryptopunkscc/astrald/node/net"
	_net "net"
)

// Conn represents a net.Conn over a TCP/IPv4 connection
type Conn struct {
	_net.Conn
	outbound bool
}

// Wrap a net.Conn from standard libraries into a astral's net.Conn
func Wrap(conn _net.Conn, outbound bool) *Conn {
	return &Conn{
		Conn:     conn,
		outbound: outbound,
	}
}

var _ net2.Conn = &Conn{}

func (conn Conn) RemoteEndpoint() net2.Endpoint {
	return net2.Endpoint{
		Net:     conn.RemoteAddr().Network(),
		Address: conn.RemoteAddr().String(),
	}
}

func (conn Conn) Outbound() bool {
	return conn.outbound
}
