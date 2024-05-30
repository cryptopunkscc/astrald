package objects

import (
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/modules"
)

type Loader struct{}

func (Loader) Load(node modules.Node, assets assets.Assets, log *log.Logger) (modules.Module, error) {
	var mod = &Module{
		node:   node,
		config: defaultConfig,
		log:    log,
	}

	mod.events.SetParent(node.Events())

	_ = assets.LoadYAML(objects.ModuleName, &mod.config)

	mod.db = assets.Database()

	err := mod.db.AutoMigrate(&dbHolding{})
	if err != nil {
		return nil, err
	}

	err = mod.node.Auth().Add(mod)
	if err != nil {
		panic(err)
	}

	mod.AddFinder(&LinkedFinder{mod: mod})

	return mod, nil
}

func init() {
	if err := modules.RegisterModule(objects.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
