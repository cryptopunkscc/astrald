package nodes

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"time"
)

const ModuleName = "nodes"

type Module interface {
	exonet.Resolver
	Accept(ctx context.Context, conn exonet.Conn) error

	ParseInfo(s string) (*NodeInfo, error)

	AddEndpoint(id.Identity, ...exonet.Endpoint) error
	RemoveEndpoint(id.Identity, ...exonet.Endpoint) error

	Endpoints(id.Identity) []exonet.Endpoint

	Peers() []id.Identity
}

// Link is an encrypted communication channel between two identities that is capable of routing queries
type Link interface {
	astral.Router
	LocalIdentity() id.Identity
	RemoteIdentity() id.Identity
	Close() error
	Done() <-chan struct{}
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
