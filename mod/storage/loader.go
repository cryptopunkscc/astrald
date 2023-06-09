package storage

import (
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/modules"
)

const ModuleName = "storage"
const configName = "storage"

type Loader struct{}

func (Loader) Load(node modules.Node, assets assets.Store, log *log.Logger) (modules.Module, error) {
	var err error
	var mod = &Module{
		node:    node,
		config:  defaultConfig,
		sources: make(map[*Source]struct{}, 0),
		log:     log,
	}

	assets.LoadYAML(configName, &mod.config)

	mod.db, err = assets.OpenDB(ModuleName)
	if err != nil {
		return nil, err
	}

	if err = mod.setupDatabase(); err != nil {
		return nil, err
	}

	adm, err := modules.Find[*admin.Module](node.Modules())
	if err == nil {
		adm.AddCommand("storage", NewAdmin(mod))
	}

	return mod, nil
}

func init() {
	if err := modules.RegisterModule(ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
