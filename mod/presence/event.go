package presence

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
)

type EventIdentityPresent struct {
	Identity id.Identity
	Endpoint net.Endpoint
}

type EventIdentityGone struct {
	Identity id.Identity
}
