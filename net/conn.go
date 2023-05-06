package net

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"io"
)

// Conn represents a raw network connection
type Conn interface {
	io.ReadWriteCloser        // Basic IO operations
	Outbound() bool           // Returns true if we are the active party, false otherwise
	LocalEndpoint() Endpoint  // Returns local network address if known, nil otherwise
	RemoteEndpoint() Endpoint // Returns the other party's network address if known, nil otherwise
}

// SecureConn represents a network connection that is encrypted and authenticated
type SecureConn interface {
	Conn
	RemoteIdentity() id.Identity // Returns the remote identity
	LocalIdentity() id.Identity  // Returns the local identity
}
