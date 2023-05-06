package tor

import (
	"github.com/cryptopunkscc/astrald/net"
	_net "net"
)

// Type check
var _ net.Conn = Conn{}

// Conn represents a network connection over Driver
type Conn struct {
	_net.Conn
	remoteEndpoint Endpoint
	outbound       bool
}

// LocalEndpoint returns an empty address, since there is no local endpoint in Driver
func (conn Conn) LocalEndpoint() net.Endpoint {
	return Endpoint{}
}

// RemoteEndpoint returns the address of the remote party
func (conn Conn) RemoteEndpoint() net.Endpoint {
	return conn.remoteEndpoint
}

// Outbound returns true if the connection is outbound
func (conn Conn) Outbound() bool {
	return conn.outbound
}

// newConn wraps a standard net.Conn into a astral's net.Conn with the addition of boundness
func newConn(conn _net.Conn, remoteAddr Endpoint, outbound bool) Conn {
	return Conn{
		Conn:           conn,
		remoteEndpoint: remoteAddr,
		outbound:       outbound,
	}
}
