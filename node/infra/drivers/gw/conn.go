package gw

import (
	"github.com/cryptopunkscc/astrald/net"
	"io"
)

// Type check
var _ net.Conn = Conn{}

// Conn represents a network connection over Astral
type Conn struct {
	io.ReadWriteCloser
	remoteAddr Endpoint
	outbound   bool
}

// LocalAddr returns an empty address, since there is no local endpoint in Tor
func (conn Conn) LocalEndpoint() net.Endpoint {
	return Endpoint{}
}

// RemoteAddr returns the address of the remote party
func (conn Conn) RemoteEndpoint() net.Endpoint {
	return conn.remoteAddr
}

// Outbound returns true if the connection is outbound
func (conn Conn) Outbound() bool {
	return conn.outbound
}

// newConn wraps an io.ReadWriteCloser with drivers.Conn interface
func newConn(rwc io.ReadWriteCloser, remoteAddr Endpoint, outbound bool) Conn {
	return Conn{
		ReadWriteCloser: rwc,
		remoteAddr:      remoteAddr,
		outbound:        outbound,
	}
}
