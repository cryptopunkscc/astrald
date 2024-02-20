package nodes

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
	"time"
)

const ModuleName = "nodes"
const DBPrefix = "nodes__"

type Module interface {
	Resolver
	AddEndpoint(nodeID id.Identity, endpoint net.Endpoint) error
	RemoveEndpoint(nodeID id.Identity, endpoint net.Endpoint) error
	Forget(nodeID id.Identity) error
}

type Resolver interface {
	Resolve(context.Context, id.Identity, *ResolveOpts) ([]net.Endpoint, error)
}

type Info struct {
	Linked      bool
	LastLinked  time.Time
	FirstLinked time.Time
}

type ResolveOpts struct {
	Network bool
	Filter  id.Filter
}

type Endpoint struct {
	Network string
	Address string
}

type Desc struct {
	Endpoints []Endpoint
}

func (Desc) Type() string {
	return "node"
}
