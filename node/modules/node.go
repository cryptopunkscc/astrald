package modules

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/node/infra"
	"github.com/cryptopunkscc/astrald/node/network"
	"github.com/cryptopunkscc/astrald/node/resolver"
	"github.com/cryptopunkscc/astrald/node/router"
	"github.com/cryptopunkscc/astrald/node/tracker"
)

// Node is a subset of node.Node that's exposed to modules
type Node interface {
	Identity() id.Identity
	Events() *events.Queue
	Infra() infra.Infra
	Network() network.Network
	Tracker() tracker.Tracker
	Modules() Modules
	Resolver() resolver.Resolver
	Router() router.Router
	AddRoute(string, net.Router) error
	RemoveRoute(string) error
	Routes() []router.QueryRoute
}
