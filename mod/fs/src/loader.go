package fs

import (
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/fs"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/modules"
)

type Loader struct{}

func (Loader) Load(node modules.Node, assets assets.Assets, log *log.Logger) (modules.Module, error) {
	var err error
	var mod = &Module{
		node:   node,
		assets: assets,
		log:    log,
		config: defaultConfig,
	}

	mod.events.SetParent(node.Events())

	_ = assets.LoadYAML(fs.ModuleName, &mod.config)

	// set up database
	mod.db = assets.Database()

	if err := mod.db.AutoMigrate(&dbLocalFile{}); err != nil {
		return nil, err
	}

	// set up services
	mod.indexer = NewIndexerService(mod)
	for _, path := range mod.config.Index {
		mod.indexer.Add(path)
	}

	mod.store, err = NewStoreService(mod)
	if err != nil {
		return nil, err
	}

	return mod, nil
}

func init() {
	if err := modules.RegisterModule(fs.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
