package tcp

import (
	"context"
	"sync"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/tcp"
	"github.com/cryptopunkscc/astrald/mod/tree"
	"github.com/cryptopunkscc/astrald/sig"
)

var _ tcp.Module = &Module{}

type Module struct {
	Deps
	config   Config
	settings Settings

	node            astral.Node
	log             *log.Logger
	ctx             *astral.Context
	configEndpoints []exonet.Endpoint
	ops             ops.Set

	mu sync.Mutex

	server             sig.Switch
	ephemeralListeners sig.Map[astral.Uint16, exonet.EphemeralListener]
}

type Settings struct {
	Listen *tree.Value[*astral.Bool] `tree:"listen"`
	Dial   *tree.Value[*astral.Bool] `tree:"dial"`
}

func (mod *Module) GetOpSet() *ops.Set {
	return &mod.ops
}

func (mod *Module) String() string {
	return tcp.ModuleName
}

func (mod *Module) acceptAll(ctx context.Context, conn exonet.Conn) (shouldStop bool, err error) {
	err = mod.Nodes.EstablishInboundLink(ctx, conn)
	if err != nil {
		return false, err
	}

	return false, nil
}

func (mod *Module) Run(ctx *astral.Context) (err error) {
	mod.ctx = ctx

	err = mod.syncConfig(ctx)
	if err != nil {
		return err
	}

	go func() {
		for v := range mod.settings.Listen.Follow(ctx) {
			mod.server.Set(ctx, v == nil || bool(*v), mod.startServer)
		}
	}()

	<-ctx.Done()

	return nil
}

func (mod *Module) ListenPort() int {
	return mod.config.ListenPort
}

func (mod *Module) endpoints() (list []*nodes.EndpointWithTTL) {
	ips, _ := mod.IP.LocalIPs()
	for _, tip := range ips {
		list = append(list, nodes.NewEndpointWithTTL(&tcp.Endpoint{
			IP:   tip,
			Port: astral.Uint16(mod.config.ListenPort),
		}, 7*24*time.Hour))
	}

	for _, e := range mod.configEndpoints {
		list = append(list, nodes.NewEndpointWithTTL(e, 7*24*time.Hour))
	}

	return list
}

func (mod *Module) syncConfig(ctx *astral.Context) error {
	if mod.config.Dial != nil {
		val := astral.Bool(*mod.config.Dial)
		if err := mod.settings.Dial.Set(ctx, &val); err != nil {
			return err
		}
	}

	if mod.config.Listen != nil {
		val := astral.Bool(*mod.config.Listen)
		if err := mod.settings.Listen.Set(ctx, &val); err != nil {
			return err
		}
	}

	return nil
}
