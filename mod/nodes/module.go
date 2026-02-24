package nodes

import (
	"context"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
)

const (
	StrategyBasic = "basic"
	StrategyTor   = "tor"
	StrategyNAT   = "nat"
)

const (
	ModuleName     = "nodes"
	DBPrefix       = "nodes__"
	ActionRelayFor = "mod.nodes.relay_for"

	CleanupGrace    = 30 * 24 * time.Hour
	CleanupInterval = 24 * time.Hour

	// query extra keys
	ExtraCallerProof   = "caller_proof"
	ExtraRelayVia      = "relay_via"
	ExtraRoutingPolicy = "routing_policy"

	// MethodMigrateSession is the query route for Phase 0 migration signaling.
	MethodMigrateSession = "nodes.migrate_session"
	// MethodResolveEndpoints is the query route for resolving endpoints of a node.
	MethodResolveEndpoints = "nodes.resolve_endpoints"
)

const ()

type Module interface {
	EstablishInboundLink(ctx context.Context, conn exonet.Conn) error
	EstablishOutboundLink(ctx context.Context, remoteID *astral.Identity, conn exonet.Conn) error

	AddEndpoint(*astral.Identity, *EndpointWithTTL) error
	RemoveEndpoint(*astral.Identity, exonet.Endpoint) error

	ResolveEndpoints(*astral.Context, *astral.Identity) (<-chan *EndpointWithTTL, error)
	AddResolver(resolver EndpointResolver)

	Peers() []*astral.Identity

	IsLinked(*astral.Identity) bool

	NewCreateStreamTask(target *astral.Identity, endpoint exonet.Endpoint) CreateStreamTask
	NewEnsureStreamTask(target *astral.Identity, strategies []string, networks []string, create bool) EnsureStreamTask
	NewCleanupEndpointsTask() CleanupEndpointsTask
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
	ResolveEndpoints(*astral.Context, *astral.Identity) (<-chan *EndpointWithTTL, error)
}

type LinkStrategy interface {
	Name() string
	Signal(ctx *astral.Context)
	Done() <-chan struct{}
}

type StrategyFactory interface {
	Build(target *astral.Identity) LinkStrategy
}
