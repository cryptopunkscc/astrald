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

	// MethodMigrateSession is the query route for Phase 0 migration signaling.
	MethodMigrateSession = "nodes.migrate_session"
)

type Module interface {
	Accept(ctx context.Context, conn exonet.Conn) error
	Connect(ctx context.Context, remoteID *astral.Identity, conn exonet.Conn) error

	AddEndpoint(*astral.Identity, exonet.Endpoint) error
	RemoveEndpoint(*astral.Identity, exonet.Endpoint) error

	ResolveEndpoints(*astral.Context, *astral.Identity) (<-chan exonet.Endpoint, error)
	AddResolver(resolver EndpointResolver)

	Peers() []*astral.Identity

	NewCreateStreamAction(target string, net string, endpoint string) CreateStreamAction
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
