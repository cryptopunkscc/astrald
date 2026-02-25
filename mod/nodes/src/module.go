package nodes

import (
	"context"
	"errors"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/crypto"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	modsecp256k1 "github.com/cryptopunkscc/astrald/mod/secp256k1"
	"github.com/cryptopunkscc/astrald/resources"
	"github.com/cryptopunkscc/astrald/sig"
)

const DefaultWorkerCount = 8
const infoPrefix = "node1"
const featureMux2 = "mux2"
const defaultPingTimeout = time.Second * 30
const activeInterval = 1 * time.Second
const pingJitter = 1 * time.Second

type NodeInfo nodes.NodeInfo

var _ nodes.Module = &Module{}

type Module struct {
	Deps
	config Config
	node   astral.Node
	log    *log.Logger
	assets resources.Resources
	db     *DB
	ctx    *astral.Context
	ops    ops.Set

	dbResolver *DBEndpointResolver
	resolvers  sig.Set[nodes.EndpointResolver]
	relays     sig.Map[astral.Nonce, *Relay]

	observedEndpoints sig.Map[string, ObservedEndpoint] // key is IP string

	peers    *Peers
	linkPool *LinkPool

	strategyFactories sig.Map[string, nodes.StrategyFactory]

	in chan *Frame

	searchCache sig.Map[string, *astral.Identity]

	privateKey *crypto.PrivateKey
}

type Relay struct {
	Caller *astral.Identity
	Target *astral.Identity
}

func (mod *Module) Run(ctx *astral.Context) error {
	mod.ctx = ctx.IncludeZone(astral.ZoneNetwork)
	<-mod.Deps.Scheduler.Ready()
	go mod.peers.frameReader(ctx)
	mod.Scheduler.Schedule(mod.NewCleanupEndpointsTask())

	<-ctx.Done()
	return nil
}

func (mod *Module) Peers() (peers []*astral.Identity) {
	return mod.peers.peers()
}

func (mod *Module) IsPeer(id *astral.Identity) bool {
	for _, peer := range mod.peers.peers() {
		if peer.IsEqual(id) {
			return true
		}
	}
	return false
}

func (mod *Module) EstablishInboundLink(ctx context.Context, conn exonet.Conn) (err error) {
	return mod.peers.EstablishInboundLink(ctx, conn)
}

func (mod *Module) EstablishOutboundLink(ctx context.Context, target *astral.Identity, conn exonet.Conn) error {
	_, err := mod.peers.EstablishOutboundLink(ctx, target, conn)
	return err
}

func (mod *Module) AddEndpoint(nodeID *astral.Identity, endpoint *nodes.EndpointWithTTL) error {
	var expiresAt *time.Time
	if endpoint.TTL != nil {
		t := time.Now().UTC().Add(time.Duration(*endpoint.TTL) * time.Second)
		expiresAt = &t
	}

	return mod.db.AddEndpoint(nodeID, endpoint.Network(), endpoint.Address(), expiresAt)
}

func (mod *Module) RemoveEndpoint(nodeID *astral.Identity, endpoint exonet.Endpoint) error {
	return mod.db.RemoveEndpoint(nodeID, endpoint.Network(), endpoint.Address())
}

// CloseStream closes a stream with the given id.
func (mod *Module) CloseStream(id astral.Nonce) error {
	streams := mod.peers.streams.Clone()
	for _, s := range streams {
		if s.id == id {
			return s.CloseWithError(errors.New("stream closed"))
		}
	}

	return errors.New("stream not found")
}

func (mod *Module) GetOpSet() *ops.Set {
	return &mod.ops
}

func (mod *Module) String() string {
	return nodes.ModuleName
}

func (mod *Module) AddResolver(resolver nodes.EndpointResolver) {
	if resolver != nil {
		mod.resolvers.Add(resolver)
	}
}

func (mod *Module) RegisterLinkStrategy(network string, factory nodes.StrategyFactory) {
	mod.strategyFactories.Set(network, factory)
}

func (mod *Module) IsLinked(identity *astral.Identity) bool {
	return mod.peers.isLinked(identity)
}

func (mod *Module) getPrivateKey() (_ *crypto.PrivateKey, err error) {
	if mod.privateKey != nil {
		return mod.privateKey, nil
	}

	mod.privateKey, err = mod.Crypto.PrivateKey(mod.ctx, modsecp256k1.FromIdentity(mod.ctx.Identity()))
	if err != nil {
		return nil, err
	}

	return mod.privateKey, nil
}

// findStreamByID returns a stream with the given local id or nil if not found.
func (mod *Module) findStreamByID(id astral.Nonce) *Stream {
	for _, s := range mod.peers.streams.Clone() {
		if s.id == id {
			return s
		}
	}
	return nil
}

func (mod *Module) createSessionMigrator(
	ctx *astral.Context,
	role migrateRole,
	ch *channel.Channel,
	peer *astral.Identity,
	sessionId astral.Nonce,
	streamId astral.Nonce,
) (migrator sessionMigrator, err error) {
	session, ok := mod.peers.sessions.Get(sessionId)
	if !ok {
		return migrator, errors.New("session not found")
	}

	return sessionMigrator{
		mod:       mod,
		sess:      session,
		role:      role,
		ch:        ch,
		local:     ctx.Identity(),
		peer:      peer,
		sessionId: sessionId,
		streamId:  streamId,
	}, nil
}
