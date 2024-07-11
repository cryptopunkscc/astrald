package network

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/infra"
	"github.com/cryptopunkscc/astrald/node/router"
)

type Node interface {
	Identity() id.Identity
	Router() router.Router
	Infra() infra.Infra
}
