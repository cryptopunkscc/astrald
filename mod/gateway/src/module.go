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
	"github.com/cryptopunkscc/astrald/mod/services"
	tcpmod "github.com/cryptopunkscc/astrald/mod/tcp"
	"github.com/cryptopunkscc/astrald/sig"
)

const NetworkName = "gw"

type Deps struct {
	Dir       dir.Module
	Exonet    exonet.Module
	Nodes     nodes.Module
	Scheduler scheduler.Module
	Services  services.Module
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

	gateways   sig.Set[*astral.Identity]
	binders    sig.Map[string, *binder]
	connectors sig.Set[*connector]

	listenEndpoints sig.Map[string, exonet.Endpoint]
}

var _ gateway.Module = &Module{}

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
		mod.addPersistentGateway(gw)
	}

	<-ctx.Done()
	for _, b := range mod.binders.Values() {
		b.Close()
	}
	for _, c := range mod.connectors.Clone() {
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

func (mod *Module) binderByIdentity(identity *astral.Identity) (*binder, bool) {
	return mod.binders.Get(identity.String())
}

func (mod *Module) binderByNonce(nonce astral.Nonce) (*binder, bool) {
	for _, b := range mod.binders.Values() {
		if b.Nonce == nonce {
			return b, true
		}
	}
	return nil, false
}

func (mod *Module) connectorByNonce(nonce astral.Nonce) (*connector, bool) {
	for _, c := range mod.connectors.Clone() {
		if c.Nonce == nonce {
			return c, true
		}
	}
	return nil, false
}

func (mod *Module) canGateway(identity *astral.Identity) bool {
	return mod.config.Gateway.Enabled
}

func (mod *Module) addPersistentGateway(gatewayID *astral.Identity) {
	mod.gateways.Add(gatewayID)
	mod.Scheduler.Schedule(mod.NewMaintainBindingTask(gatewayID, mod.config.Visibility))
}

func (mod *Module) String() string {
	return gateway.ModuleName
}
