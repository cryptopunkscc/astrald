package presence

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
)

// Presence holds information about an identity present on the network
type Presence struct {
	Identity id.Identity
	Endpoint net.Endpoint
	Present  bool
}
