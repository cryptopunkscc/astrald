package tor

import (
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/tree"
	"golang.org/x/net/proxy"
)

const defaultListenPort = 1791

type Deps struct {
	Nodes  nodes.Module
	Exonet exonet.Module
	Tree   tree.Module
}

type Settings struct {
	Listen *tree.Value[*astral.Bool] `tree:"listen"`
	Dial   *tree.Value[*astral.Bool] `tree:"dial"`
}

type Module struct {
	Deps
	config   Config
	settings Settings

	node   astral.Node
	assets assets.Assets
	log    *log.Logger
	ctx    *astral.Context
	proxy  proxy.ContextDialer
	server *Server

	serverMu     sync.Mutex
	serverCancel func()
}

func (mod *Module) Run(ctx *astral.Context) error {
	mod.ctx = ctx

	if err := mod.loadSettings(ctx); err != nil {
		return err
	}

	if v := mod.settings.Listen.Get(); v == nil || *v {
		mod.startServer()
	}

	<-ctx.Done()

	return nil
}

func (mod *Module) loadSettings(ctx *astral.Context) error {
	falseValue := astral.Bool(false)

	if mod.config.Dial != nil && !*mod.config.Dial {
		if err := mod.settings.Dial.Set(ctx, &falseValue); err != nil {
			return err
		}
	}

	if mod.config.Listen != nil && !*mod.config.Listen {
		if err := mod.settings.Listen.Set(ctx, &falseValue); err != nil {
			return err
		}
	}

	return nil
}
