package nodes

import (
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/mod/exonet"
)

type EventNewEndpoint struct {
	Identity id.Identity
	Endpoint exonet.Endpoint
}
