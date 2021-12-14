package tor

import (
	"github.com/cryptopunkscc/astrald/infra"
	"net"
)

// Type check
var _ infra.Conn = Conn{}

// Conn represents a network connection over Tor
type Conn struct {
	net.Conn
	remoteAddr Addr
	outbound   bool
}

// LocalAddr returns an empty address, since there is no local endpoint in Tor
func (conn Conn) LocalAddr() infra.Addr {
	return Addr{}
}

// RemoteAddr returns the address of the remote party
func (conn Conn) RemoteAddr() infra.Addr {
	return conn.remoteAddr
}

// Outbound returns true if the connection is outbound
func (conn Conn) Outbound() bool {
	return conn.outbound
}

// newConn wraps a standard net.Conn into a astral's net.Conn with the addition of boundness
func newConn(conn net.Conn, remoteAddr Addr, outbound bool) Conn {
	return Conn{
		Conn:       conn,
		remoteAddr: remoteAddr,
		outbound:   outbound,
	}
}
