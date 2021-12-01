package astral

import (
	"github.com/cryptopunkscc/astrald/infra"
	"io"
)

// Type check
var _ infra.Conn = Conn{}

// Conn represents a network connection over Astral
type Conn struct {
	io.ReadWriteCloser
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

// newConn wraps an io.ReadWriteCloser with infra.Conn interface
func newConn(rwc io.ReadWriteCloser, remoteAddr Addr, outbound bool) Conn {
	return Conn{
		ReadWriteCloser: rwc,
		remoteAddr:      remoteAddr,
		outbound:        outbound,
	}
}
