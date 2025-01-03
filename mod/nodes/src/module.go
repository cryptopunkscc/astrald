package nodes

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/resources"
	"github.com/cryptopunkscc/astrald/sig"
	"github.com/jxskiss/base62"
	"gorm.io/gorm"
	"io"
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
	Admin   admin.Module
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
	db     *gorm.DB

	peers    *Peers
	provider *Provider

	in chan *Frame

	searchCache sig.Map[string, *astral.Identity]
}

func (mod *Module) Run(ctx context.Context) error {
	go mod.peers.frameReader(ctx)
	<-ctx.Done()
	return nil
}

func (mod *Module) Peers() (peers []*astral.Identity) {
	return mod.peers.peers()
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

func (mod *Module) ResolveEndpoints(ctx context.Context, identity *astral.Identity) ([]exonet.Endpoint, error) {
	return mod.Endpoints(identity), nil
}

func (mod *Module) RouteQuery(ctx context.Context, q *astral.Query, w io.WriteCloser) (rw io.WriteCloser, err error) {
	if s, ok := q.Extra.Get("origin"); ok && s == "network" {
		return query.RouteNotFound(mod)
	}

	var relayID = q.Target
	var callerProof astral.Object

	useRelay := false

	if !q.Caller.IsEqual(mod.node.Identity()) {
		useRelay = true
		if v, ok := q.Extra.Get(nodes.ExtraCallerProof); ok {
			callerProof = v.(astral.Object)
		}
	}

	if v, ok := q.Extra.Get(nodes.ExtraRelayVia); ok {
		relayID = v.(*astral.Identity)
		useRelay = true
	}

	if useRelay {
		if callerProof != nil {
			err = mod.Objects.Push(ctx, nil, relayID, callerProof)
			if err != nil {
				mod.log.Errorv(1, "cannot push proof: %v", err)
			}
		}

		err = mod.on(relayID).Relay(ctx, q.Nonce, q.Caller, q.Target)
		if err != nil {
			return query.RouteNotFound(mod, err)
		}

		if !q.Target.IsEqual(relayID) {
			q = &astral.Query{
				Nonce:  q.Nonce,
				Caller: q.Caller,
				Target: relayID,
				Query:  q.Query,
				Extra:  *q.Extra.Copy(),
			}
		}
	}

	return mod.peers.RouteQuery(ctx, q, w)
}

func (mod *Module) on(providerID *astral.Identity) *Consumer {
	return NewConsumer(mod, providerID)
}

func (mod *Module) String() string {
	return nodes.ModuleName
}
