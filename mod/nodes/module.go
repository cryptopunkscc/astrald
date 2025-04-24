package nodes

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
)

const ModuleName = "nodes"

type Module interface {
	Accept(ctx context.Context, conn exonet.Conn) error

	ParseInfo(s string) (*NodeInfo, error)

	AddEndpoint(*astral.Identity, exonet.Endpoint) error
	RemoveEndpoint(*astral.Identity, exonet.Endpoint) error

	ResolveEndpoints(*astral.Context, *astral.Identity) (<-chan exonet.Endpoint, error)
	AddResolver(resolver EndpointResolver)

	Peers() []*astral.Identity
}

// Link is an encrypted communication channel between two identities that is capable of routing queries
type Link interface {
	astral.Router
	LocalIdentity() *astral.Identity
	RemoteIdentity() *astral.Identity
	Close() error
	Done() <-chan struct{}
}

type NodeInfo struct {
	Identity  *astral.Identity
	Alias     string
	Endpoints []exonet.Endpoint
}

type EndpointResolver interface {
	ResolveEndpoints(*astral.Context, *astral.Identity) (<-chan exonet.Endpoint, error)
}
