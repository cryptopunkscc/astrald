package nodes

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/resources"
	"github.com/cryptopunkscc/astrald/sig"
	"github.com/jxskiss/base62"
	"strings"
	"time"
)

const DefaultWorkerCount = 8
const infoPrefix = "node1"
const featureMux2 = "mux2"
const defaultPingTimeout = time.Second * 30

type NodeInfo nodes.NodeInfo

var _ nodes.Module = &Module{}

type Deps struct {
	Auth    auth.Module
	Dir     dir.Module
	Exonet  exonet.Module
	Keys    keys.Module
	Objects objects.Module
}

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

	peers *Peers

	in chan *Frame

	searchCache sig.Map[string, *astral.Identity]
}

type Relay struct {
	Caller *astral.Identity
	Target *astral.Identity
}

func (mod *Module) Run(ctx *astral.Context) error {
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

func (mod *Module) ParseInfo(s string) (*nodes.NodeInfo, error) {
	trimmed := strings.TrimPrefix(s, infoPrefix)
	data, err := base62.DecodeString(trimmed)
	if err != nil {
		return nil, err
	}

	info, err := (&InfoEncoder{mod}).Unpack(data)
	if err != nil {
		return nil, err
	}

	return (*nodes.NodeInfo)(info), nil
}

func (mod *Module) InfoString(info *nodes.NodeInfo) string {
	packed, err := (&InfoEncoder{mod}).Pack((*NodeInfo)(info))
	if err != nil {
		return ""
	}

	return infoPrefix + base62.EncodeToString(packed)
}

func (mod *Module) HasEndpoints(nodeID *astral.Identity) (has bool) {
	return mod.db.HasEndpoints(nodeID)
}

func (mod *Module) AddEndpoint(nodeID *astral.Identity, endpoint exonet.Endpoint) error {
	return mod.db.AddEndpoint(nodeID, endpoint.Network(), endpoint.Address())
}

func (mod *Module) RemoveEndpoint(nodeID *astral.Identity, endpoint exonet.Endpoint) error {
	return mod.db.RemoveEndpoint(nodeID, endpoint.Network(), endpoint.Address())
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
