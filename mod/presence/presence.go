package presence

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
)

// Presence holds information about an identity present on the network
type Presence struct {
	Identity id.Identity
	Addr     infra.Addr
	Present  bool
}
