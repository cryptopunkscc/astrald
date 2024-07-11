package nodes

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
)

type EventNewEndpoint struct {
	Identity id.Identity
	Endpoint net.Endpoint
}
