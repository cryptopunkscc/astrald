package query

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
)

type Link interface {
	Router
	LocalIdentity() id.Identity
	RemoteIdentity() id.Identity
	Transport() net.SecureConn
}

type Linker interface {
	Link() Link
}
