package tor

import (
	"net"

	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/tor"
)

// Type check
var _ exonet.Conn = &Conn{}

// Conn represents a network connection over Driver
type Conn struct {
	net.Conn
	remoteEndpoint *tor.Endpoint
	outbound       bool
}

// LocalEndpoint returns an empty address, since there is no local endpoint in Driver
func (conn *Conn) LocalEndpoint() exonet.Endpoint {
	return &tor.Endpoint{}
}

// RemoteEndpoint returns the address of the remote party
func (conn *Conn) RemoteEndpoint() exonet.Endpoint {
	if conn.remoteEndpoint == nil {
		return nil // return a generic nil instead of (*tor.Endpoint)(nil)
	}
	return conn.remoteEndpoint
}

// Outbound returns true if the connection is outbound
func (conn *Conn) Outbound() bool {
	return conn.outbound
}

// newConn wraps a standard net.Conn into a astral's net.Conn with the addition of boundness
func newConn(conn net.Conn, remoteEndpoint *tor.Endpoint, outbound bool) exonet.Conn {
	return &Conn{
		Conn:           conn,
		remoteEndpoint: remoteEndpoint,
		outbound:       outbound,
	}
}
