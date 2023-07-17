package route

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
)

type replaceIdentity struct {
	net.SecureConn
	remoteIdentity id.Identity
}

func (r replaceIdentity) Identity() id.Identity {
	return r.remoteIdentity
}
