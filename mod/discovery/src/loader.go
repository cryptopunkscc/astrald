package discovery

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/lib/routers"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/discovery"
)

type Loader struct{}

func (Loader) Load(node astral.Node, assets assets.Assets, log *log.Logger) (core.Module, error) {
	var err error
	var mod = &Module{
		node:       node,
		config:     defaultConfig,
		log:        log,
		assets:     assets,
		PathRouter: routers.NewPathRouter(node.Identity(), false),
	}

	mod.events.SetParent(node.Events())

	_ = assets.LoadYAML(discovery.ModuleName, &mod.config)

	return mod, err
}

func init() {
	if err := core.RegisterModule(discovery.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
