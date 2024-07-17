package node

import (
	"github.com/cryptopunkscc/astrald/events"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/lib/routers"
	"github.com/cryptopunkscc/astrald/net"
	"time"
)

// Node defines the overall structure of an astral node
type Node interface {
	Identity() id.Identity
	Events() *events.Queue
	Auth() AuthEngine
	Modules() ModuleEngine
	Resolver() ResolverEngine
	Router() net.Router

	LocalRouter() routers.LocalRouter
}

type ActiveLink interface {
	net.Link
	ID() int
	AddedAt() time.Time
}

type LinkSet interface {
	Add(l net.Link) (ActiveLink, error)
	ByRemoteIdentity(identity id.Identity) LinkSet
	Find(id int) (ActiveLink, error)
	All() []ActiveLink
	Count() int
}

type Resolver interface {
	Resolve(s string) (id.Identity, error)
	DisplayName(identity id.Identity) string
}

type ResolverEngine interface {
	Resolver
	AddResolver(Resolver) error
}

type Router interface {
	net.Router
	AddRoute(caller id.Identity, target id.Identity, router net.Router, priority int) error
	RemoveRoute(caller id.Identity, target id.Identity, router net.Router) error
	Routes() []Route
}

type Route struct {
	Caller   id.Identity
	Target   id.Identity
	Router   net.Router
	Priority int
}
