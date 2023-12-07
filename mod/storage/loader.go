package storage

import (
	"github.com/cryptopunkscc/astrald/log"
	storage "github.com/cryptopunkscc/astrald/mod/storage/api"
	_assets "github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/modules"
)

type Loader struct{}

func (Loader) Load(node modules.Node, assets _assets.Store, log *log.Logger) (modules.Module, error) {
	var err error
	var mod = &Module{
		node:   node,
		config: defaultConfig,
		log:    log,
	}
	mod.access = NewAccessManager(mod)
	mod.data = NewDataManager(mod)

	mod.events.SetParent(node.Events())

	_ = assets.LoadYAML(storage.ModuleName, &mod.config)

	mod.db, err = assets.OpenDB(storage.ModuleName)
	if err != nil {
		return nil, err
	}

	if err = mod.setupDatabase(); err != nil {
		return nil, err
	}

	return mod, nil
}

func init() {
	if err := modules.RegisterModule(storage.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
