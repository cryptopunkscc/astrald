package tor

import (
	"github.com/cryptopunkscc/astrald/mod/exonet"
	_net "net"
)

// Type check
var _ exonet.Conn = &Conn{}

// Conn represents a network connection over Driver
type Conn struct {
	_net.Conn
	remoteEndpoint *Endpoint
	outbound       bool
}

// LocalEndpoint returns an empty address, since there is no local endpoint in Driver
func (conn *Conn) LocalEndpoint() exonet.Endpoint {
	return nil
}

// RemoteEndpoint returns the address of the remote party
func (conn *Conn) RemoteEndpoint() exonet.Endpoint {
	return conn.remoteEndpoint
}

// Outbound returns true if the connection is outbound
func (conn *Conn) Outbound() bool {
	return conn.outbound
}

// newConn wraps a standard net.Conn into a astral's net.Conn with the addition of boundness
func newConn(conn _net.Conn, remoteEndpoint *Endpoint, outbound bool) exonet.Conn {
	return &Conn{
		Conn:           conn,
		remoteEndpoint: remoteEndpoint,
		outbound:       outbound,
	}
}
