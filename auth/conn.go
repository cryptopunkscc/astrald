package auth

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
)

// Conn is an authenticated connection. It augments net.Conn with remote identity.
type Conn interface {
	net.Conn                      // auth.Conn is an extension to the net.Conn interface
	RemoteIdentity() *id.Identity // Returns the remote identity
	LocalIdentity() *id.Identity  // Returns the local identity
}
