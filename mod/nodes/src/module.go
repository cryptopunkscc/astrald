package nodes

import (
	"context"
	"errors"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/lib/routing"
	"github.com/cryptopunkscc/astrald/mod/crypto"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/gateway"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	modsecp256k1 "github.com/cryptopunkscc/astrald/mod/secp256k1"
	"github.com/cryptopunkscc/astrald/resources"
	"github.com/cryptopunkscc/astrald/sig"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
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
	router routing.OpRouter

	dbResolver *DBEndpointResolver
	resolvers  sig.Set[nodes.EndpointResolver]

	observedEndpoints sig.Map[string, ObservedEndpoint] // key is IP string

	peers    *Peers
	linkPool *LinkPool

	strategyFactories sig.Map[string, nodes.StrategyFactory]
	upgraders         sig.Map[string, *sig.Switch]

	searchCache sig.Map[string, *astral.Identity]

	privateKey *crypto.PrivateKey
}

func (mod *Module) Run(ctx *astral.Context) error {
	mod.ctx = ctx.IncludeZone(astral.ZoneNetwork)
	<-mod.Deps.Scheduler.Ready()
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
	for _, s := range mod.linkPool.Links().Clone() {
		if s.id == id {
			return s.CloseWithError(errors.New("stream closed"))
		}
	}

	return errors.New("stream not found")
}

func (mod *Module) Router() astral.Router {
	return &mod.router
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
	return mod.linkPool.SelectLinkWith(identity) != nil
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
func (mod *Module) findStreamByID(id astral.Nonce) *Link {
	for _, s := range mod.linkPool.Links().Clone() {
		if s.id == id {
			return s
		}
	}
	return nil
}

func (mod *Module) findStreamBySessionNonce(nonce astral.Nonce) *Link {
	for _, s := range mod.linkPool.Links().Clone() {
		if _, ok := s.Mux.sessions.Get(nonce); ok {
			return s
		}
	}
	return nil
}

func (mod *Module) GetLinkNegotiator() (*muxLinkNegotiator, error) {
	privKey, err := mod.getPrivateKey()
	if err != nil {
		return nil, err
	}
	return &muxLinkNegotiator{
		mod:        mod,
		privateKey: secp256k1.PrivKeyFromBytes(privKey.Key),
		features:   []string{featureMux2},
	}, nil
}

func (mod *Module) reflectLink(s *Link) (err error) {
	if s.outbound || s.RemoteEndpoint() == nil {
		return
	}
	if _, ok := s.RemoteEndpoint().(*gateway.Endpoint); ok {
		return
	}
	err = mod.Objects.Push(mod.ctx, s.RemoteIdentity(),
		&nodes.ObservedEndpointMessage{
			Endpoint: s.RemoteEndpoint(),
		})
	if err != nil {
		mod.log.Errorv(2, "Objects.Push(%v, %v): %v", s.RemoteIdentity(), s.RemoteEndpoint(), err)
	} else {
		mod.log.Logv(2, "reflected endpoint %v to %v", s.RemoteEndpoint(), s.RemoteIdentity())
	}
	return
}
