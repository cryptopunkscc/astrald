package discovery

import (
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/discovery"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/modules"
)

type Loader struct{}

func (Loader) Load(node modules.Node, assets assets.Assets, log *log.Logger) (modules.Module, error) {
	var err error
	var mod = &Module{
		node:   node,
		config: defaultConfig,
		log:    log,
		assets: assets,
	}

	mod.events.SetParent(node.Events())

	_ = assets.LoadYAML(discovery.ModuleName, &mod.config)

	return mod, err
}

func init() {
	if err := modules.RegisterModule(discovery.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
