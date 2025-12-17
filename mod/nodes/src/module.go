package nodes

import (
	"context"
	"errors"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/resources"
	"github.com/cryptopunkscc/astrald/sig"
)

const DefaultWorkerCount = 8
const infoPrefix = "node1"
const featureMux2 = "mux2"
const defaultPingTimeout = time.Second * 30

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
	ops    shell.Scope

	dbResolver *DBEndpointResolver
	resolvers  sig.Set[nodes.EndpointResolver]
	relays     sig.Map[astral.Nonce, *Relay]

	observedEndpoints sig.Map[string, ObservedEndpoint] // key is IP string

	peers *Peers

	in chan *Frame

	searchCache sig.Map[string, *astral.Identity]
}

type Relay struct {
	Caller *astral.Identity
	Target *astral.Identity
}

func (mod *Module) Run(ctx *astral.Context) error {
	mod.ctx = ctx.IncludeZone(astral.ZoneNetwork)
	<-mod.Deps.Scheduler.Ready()
	go mod.peers.frameReader(ctx)

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

func (mod *Module) Accept(ctx context.Context, conn exonet.Conn) (err error) {
	return mod.peers.Accept(ctx, conn)
}

func (mod *Module) Connect(ctx context.Context, remoteID *astral.Identity, conn exonet.Conn) (err error) {
	_, err = mod.peers.Connect(ctx, remoteID, conn)
	return
}

func (mod *Module) AddEndpoint(nodeID *astral.Identity, endpoint exonet.Endpoint) error {
	return mod.db.AddEndpoint(nodeID, endpoint.Network(), endpoint.Address())
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

func (mod *Module) on(providerID *astral.Identity) *Consumer {
	return NewConsumer(mod, providerID)
}

func (mod *Module) Scope() *shell.Scope {
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

func (mod *Module) IsLinked(identity *astral.Identity) bool {
	return mod.peers.isLinked(identity)
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
	ch *astral.Channel,
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
