package storage

import (
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/storage"
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

	_ = assets.LoadYAML(storage.ModuleName, &mod.config)

	mod.db = assets.Database()

	return mod, nil
}

func init() {
	if err := modules.RegisterModule(storage.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
