package presence

import (
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/modules"
)

const ModuleName = "presence"

type Loader struct{}

func (Loader) Load(node modules.Node, assets assets.Assets, log *log.Logger) (modules.Module, error) {
	mod := &Module{
		node:   node,
		config: defaultConfig,
		log:    log,
	}
	mod.events.SetParent(node.Events())
	mod.discover = NewDiscoverService(mod)
	mod.announce = &AnnounceService{Module: mod}

	_ = assets.LoadYAML(ModuleName, &mod.config)

	return mod, nil
}

func init() {
	if err := modules.RegisterModule(ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
