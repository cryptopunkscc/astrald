package modules

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/authorizer"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/node/infra"
	"github.com/cryptopunkscc/astrald/node/network"
	"github.com/cryptopunkscc/astrald/node/resolver"
	"github.com/cryptopunkscc/astrald/node/router"
)

// Node is a subset of node.Node that's exposed to modules
type Node interface {
	Identity() id.Identity
	Events() *events.Queue
	Infra() infra.Infra
	Network() network.Network
	Auth() authorizer.AuthSet
	Modules() Modules
	Resolver() resolver.ResolveEngine
	Router() router.Router
	LocalRouter() router.LocalRouter
}
