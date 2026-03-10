package gateway

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/lib/routers"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	ipmod "github.com/cryptopunkscc/astrald/mod/ip"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	tcpmod "github.com/cryptopunkscc/astrald/mod/tcp"
	"github.com/cryptopunkscc/astrald/sig"
)

const NetworkName = "gw"

type Deps struct {
	Dir    dir.Module
	Exonet exonet.Module
	Nodes  nodes.Module
	TCP    tcpmod.Module
	IP     ipmod.Module
}

type Module struct {
	Deps
	*routers.PathRouter

	ops    ops.Set
	config Config
	node   astral.Node
	log    *log.Logger
	ctx    *astral.Context

	binders    sig.Set[*Binder]
	connecting sig.Set[*Connecting]

	listenEndpoints sig.Map[string, exonet.Endpoint]
}

func (mod *Module) GetOpSet() *ops.Set {
	return &mod.ops
}

func (mod *Module) Run(ctx *astral.Context) error {
	mod.ctx = ctx.IncludeZone(astral.ZoneNetwork)

	mod.startServers(mod.ctx)

	for _, gatewayStr := range mod.config.Gateways {
		identity, err := mod.Dir.ResolveIdentity(gatewayStr)
		if err != nil {
			mod.log.Error("resolve gateway %v: %v", gatewayStr, err)
			continue
		}

		go mod.bindToGateway(mod.ctx, identity, mod.config.Visibility)
	}

	<-ctx.Done()
	return nil
}

func (mod *Module) Endpoints() []exonet.Endpoint {
	var list = make([]exonet.Endpoint, 0)

	return list
}

func (mod *Module) getGatewayEndpoint(ctx *astral.Context, network string) (endpoint exonet.Endpoint, err error) {
	endpoint, ok := mod.listenEndpoints.Get(network)
	if !ok {
		// fixme: return public error (no gateway endpoint available)
		return
	}
	return endpoint, nil
}
func (mod *Module) String() string {
	return ModuleName
}
