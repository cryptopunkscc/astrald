package gateway

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/lib/routers"
	"github.com/cryptopunkscc/astrald/mod/gateway"
)

type Loader struct{}

func (Loader) Load(node astral.Node, assets assets.Assets, log *log.Logger) (core.Module, error) {
	mod := &Module{
		node:       node,
		log:        log,
		PathRouter: routers.NewPathRouter(node.Identity(), false),
		config:     defaultConfig,
	}

	_ = assets.LoadYAML(gateway.ModuleName, &mod.config)

	return mod, nil
}

func init() {
	if err := core.RegisterModule(gateway.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
