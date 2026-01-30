package tcp

import (
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/tcp"
	"github.com/cryptopunkscc/astrald/mod/tree"
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

	serverMu     sync.Mutex
	serverCancel func()
}

type Settings struct {
	Listen *tree.Value[*astral.Bool] `tree:"listen"`
	Dial   *tree.Value[*astral.Bool] `tree:"dial"`
}

func (mod *Module) Run(ctx *astral.Context) (err error) {
	mod.ctx = ctx

	err = mod.loadSettings(ctx)
	if err != nil {
		return err
	}

	go func() {
		for v := range mod.settings.Listen.Follow(ctx) {
			mod.switchServer(v)
		}
	}()

	<-ctx.Done()

	return nil
}

func (mod *Module) ListenPort() int {
	return mod.config.ListenPort
}

func (mod *Module) endpoints() (list []exonet.Endpoint) {
	ips, _ := mod.IP.LocalIPs()
	for _, tip := range ips {
		e := tcp.Endpoint{
			IP:   tip,
			Port: astral.Uint16(mod.config.ListenPort),
		}

		list = append(list, &e)
	}

	return list
}

func (mod *Module) loadSettings(ctx *astral.Context) (err error) {
	falseValue := astral.Bool(false)
	if mod.config.Dial != nil && !*mod.config.Dial {
		err = mod.settings.Dial.Set(ctx, &falseValue)
		if err != nil {
			return err
		}
	}

	if mod.config.Listen != nil && !*mod.config.Listen {
		err = mod.settings.Listen.Set(ctx, &falseValue)
		if err != nil {
			return err
		}
	}

	return nil
}
