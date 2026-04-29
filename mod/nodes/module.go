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
	ActionRelayFor = "mod.nodes.relay_for_action" // equals RelayForAction{}.ObjectType()

	CleanupGrace    = 30 * 24 * time.Hour
	CleanupInterval = 24 * time.Hour

	// query extra keys
	ExtraCallerProof   = "caller_proof"
	ExtraRelayVia      = "relay_via"
	ExtraRoutingPolicy = "routing_policy"

	// MethodResolveEndpoints is the query route for resolving endpoints of a node.
	MethodResolveEndpoints = "nodes.resolve_endpoints"
	// MethodNodeOpenRelay is the query route for opening a relay channel.
	MethodNodeOpenRelay = "nodes.node_open_relay"

	// MethodMigrateSession is the query route for session migration.
	MethodMigrateSession = "nodes.migrate_session"

	// DefaultBufferSize is the default buffer size for session I/O.
	DefaultBufferSize = 4 * 1024 * 1024
	MaxDataFrameSize  = 8192
)

type Module interface {
	EstablishInboundLink(ctx context.Context, conn exonet.Conn) error
	EstablishOutboundLink(ctx context.Context, remoteID *astral.Identity, conn exonet.Conn) error

	AddEndpoint(*astral.Identity, *EndpointWithTTL) error
	RemoveEndpoint(*astral.Identity, exonet.Endpoint) error

	UpdateNodeEndpoints(ctx *astral.Context, resolver *astral.Identity, identity *astral.Identity) error
	ResolveEndpoints(*astral.Context, *astral.Identity) (<-chan *EndpointWithTTL, error)
	AddResolver(resolver EndpointResolver)

	Peers() []*astral.Identity

	IsLinked(*astral.Identity) bool

	NewCreateStreamTask(target *astral.Identity, endpoint exonet.Endpoint) CreateStreamTask
	NewEnsureStreamTask(target *astral.Identity, strategies []string, networks []string, forceNew bool) EnsureStreamTask
	NewCleanupEndpointsTask() CleanupEndpointsTask
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
