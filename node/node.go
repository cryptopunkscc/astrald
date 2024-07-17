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
	Auth() AuthEngine
	Resolver() ResolverEngine
}

type AuthEngine interface {
	Authorize(id id.Identity, action string, args ...any) bool
	Add(Authorizer) error
	Remove(Authorizer) error
}

type Authorizer interface {
	Authorize(id id.Identity, action string, args ...any) bool
}

type Resolver interface {
	Resolve(s string) (id.Identity, error)
	DisplayName(identity id.Identity) string
}

type ResolverEngine interface {
	Resolver
	AddResolver(Resolver) error
}
