package network

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/infra"
	"github.com/cryptopunkscc/astrald/node/router"
	"github.com/cryptopunkscc/astrald/node/tracker"
)

type Node interface {
	Identity() id.Identity
	Router() router.Router
	Infra() infra.Infra
	Tracker() tracker.Tracker
}
