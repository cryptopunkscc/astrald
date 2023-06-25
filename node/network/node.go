package network

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/infra"
	"github.com/cryptopunkscc/astrald/node/tracker"
)

type Node interface {
	Identity() id.Identity
	Router() net.Router
	Infra() infra.Infra
	Tracker() tracker.Tracker
}
