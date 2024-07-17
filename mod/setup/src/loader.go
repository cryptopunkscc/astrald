package setup

import (
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/lib/routers"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/setup"
	"github.com/cryptopunkscc/astrald/node"
)

type Loader struct{}

func (Loader) Load(node node.Node, assets assets.Assets, log *log.Logger) (core.Module, error) {
	var err error
	var mod = &Module{
		node:       node,
		log:        log,
		assets:     assets,
		PathRouter: routers.NewPathRouter(node.Identity(), false),
	}

	_ = assets.LoadYAML(setup.ModuleName, &mod.config)

	return mod, err
}

func init() {
	if err := core.RegisterModule(setup.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
