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
	Link(context.Context, id.Identity, LinkOpts) (net.Link, error)
	AcceptLink(ctx context.Context, conn net.Conn) (net.Link, error)
	InitLink(ctx context.Context, conn net.Conn, remoteID id.Identity) (net.Link, error)

	ParseInfo(s string) (*NodeInfo, error)

	AddEndpoint(id.Identity, ...net.Endpoint) error
	RemoveEndpoint(id.Identity, ...net.Endpoint) error

	Endpoints(id.Identity) []net.Endpoint
}

type Resolver interface {
	Resolve(context.Context, id.Identity, *ResolveOpts) ([]net.Endpoint, error)
}

type NodeInfo struct {
	Identity  id.Identity
	Alias     string
	Endpoints []net.Endpoint
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

type LinkOpts struct {
	Endpoints []net.Endpoint
	Workers   int
}
