package gateway

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/lib/routers"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/gateway"
	ipmod "github.com/cryptopunkscc/astrald/mod/ip"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/scheduler"
	tcpmod "github.com/cryptopunkscc/astrald/mod/tcp"
	"github.com/cryptopunkscc/astrald/sig"
)

const NetworkName = "gw"

type Deps struct {
	Dir       dir.Module
	Exonet    exonet.Module
	Nodes     nodes.Module
	Scheduler scheduler.Module
	TCP       tcpmod.Module
	IP        ipmod.Module
}

type Module struct {
	Deps
	*routers.PathRouter

	ops    ops.Set
	config Config
	node   astral.Node
	log    *log.Logger
	ctx    *astral.Context

	binders    sig.Map[string, *client]
	connecting sig.Set[*client]

	listenEndpoints sig.Map[string, exonet.Endpoint]
}

func (mod *Module) GetOpSet() *ops.Set {
	return &mod.ops
}

func (mod *Module) Run(ctx *astral.Context) error {
	mod.ctx = ctx.IncludeZone(astral.ZoneNetwork)

	err := mod.AddRoute(gateway.MethodRoute+".*", routers.Func(mod.routeQuery))
	if err != nil {
		return err
	}

	if mod.config.Gateway.Enabled {
		mod.startServers(mod.ctx)
	}

	<-mod.Scheduler.Ready()
	for _, gw := range mod.config.Gateways {
		mod.Scheduler.Schedule(mod.NewMaintainBindingTask(gw, mod.config.Visibility))
	}

	<-ctx.Done()

	for _, c := range mod.binders.Values() {
		c.Close()
	}
	for _, c := range mod.connecting.Clone() {
		c.Close()
	}

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

func (mod *Module) canGateway(identity *astral.Identity) bool {
	return mod.config.Gateway.Enabled
}

func (mod *Module) String() string {
	return gateway.ModuleName
}
