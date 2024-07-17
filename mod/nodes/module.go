package nodes

import (
	"context"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/net"
	"time"
)

const ModuleName = "nodes"
const DBPrefix = "nodes__"

type Module interface {
	exonet.Resolver
	Link(context.Context, id.Identity, LinkOpts) (net.Link, error)
	AcceptLink(ctx context.Context, conn exonet.Conn) (net.Link, error)
	InitLink(ctx context.Context, conn exonet.Conn, remoteID id.Identity) (net.Link, error)

	ParseInfo(s string) (*NodeInfo, error)

	AddEndpoint(id.Identity, ...exonet.Endpoint) error
	RemoveEndpoint(id.Identity, ...exonet.Endpoint) error

	Endpoints(id.Identity) []exonet.Endpoint

	Peers() []id.Identity
}

type NodeInfo struct {
	Identity  id.Identity
	Alias     string
	Endpoints []exonet.Endpoint
}

type Info struct {
	Linked      bool
	LastLinked  time.Time
	FirstLinked time.Time
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
	Endpoints []exonet.Endpoint
	Workers   int
}
