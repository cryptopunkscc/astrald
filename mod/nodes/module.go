package nodes

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
)

const (
	ModuleName       = "nodes"
	DBPrefix         = "nodes__"
	ActionRelayFor   = "mod.nodes.relay_for"
	ExtraCallerProof = "caller_proof"
	ExtraRelayVia    = "relay_via"
)

type Module interface {
	Accept(ctx context.Context, conn exonet.Conn) error

	AddEndpoint(*astral.Identity, exonet.Endpoint) error
	RemoveEndpoint(*astral.Identity, exonet.Endpoint) error

	ResolveEndpoints(*astral.Context, *astral.Identity) (<-chan exonet.Endpoint, error)
	AddResolver(resolver EndpointResolver)

	ResolveServices(*astral.Context, *astral.Identity) []*ServiceTTL
	AddServiceResolver(resolver ServiceResolver)

	Peers() []*astral.Identity
	Services() ServiceQuery
}

// Link is an encrypted communication channel between two identities that is capable of routing queries
type Link interface {
	astral.Router
	LocalIdentity() *astral.Identity
	RemoteIdentity() *astral.Identity
	Close() error
	Done() <-chan struct{}
}

type EndpointResolver interface {
	ResolveEndpoints(*astral.Context, *astral.Identity) (<-chan exonet.Endpoint, error)
}

// ServiceResolver returns a list of services provided to the identity
type ServiceResolver interface {
	ResolveServices(context *astral.Context, identity *astral.Identity) []*ServiceTTL
}

type ServiceQuery interface {
	Find() []*Service
	ByName(name ...string) ServiceQuery
	ByNodeID(id ...*astral.Identity) ServiceQuery
}
