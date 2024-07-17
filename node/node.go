package node

import (
	"github.com/cryptopunkscc/astrald/events"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/net"
)

// Node defines the overall structure of an astral node
type Node interface {
	Identity() id.Identity
	Router() net.Router
	Events() *events.Queue
	Resolver() ResolverEngine
}

type Resolver interface {
	Resolve(s string) (id.Identity, error)
	DisplayName(identity id.Identity) string
}

type ResolverEngine interface {
	Resolver
	AddResolver(Resolver) error
}
