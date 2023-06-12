package modules

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/node/infra"
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/node/network"
	"github.com/cryptopunkscc/astrald/node/resolver"
	"github.com/cryptopunkscc/astrald/node/services"
	"github.com/cryptopunkscc/astrald/node/tracker"
)

// Node is a subset of node.Node that's exposed to modules
type Node interface {
	Identity() id.Identity
	Query(ctx context.Context, remoteID id.Identity, query string) (link.BasicConn, error)
	Events() *events.Queue
	Infra() infra.Infra
	Network() network.Network
	Tracker() tracker.Tracker
	Services() services.Services
	Modules() Modules
	Resolver() resolver.Resolver
}
