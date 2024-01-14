package keys

import (
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/modules"
)

type Loader struct{}

func (Loader) Load(node modules.Node, assets assets.Store, log *log.Logger) (modules.Module, error) {
	var err error
	var mod = &Module{
		node:   node,
		log:    log,
		assets: assets,
	}

	_ = assets.LoadYAML(keys.ModuleName, &mod.config)

	mod.db, err = mod.assets.OpenDB(keys.ModuleName)
	if err != nil {
		return nil, err
	}

	err = mod.db.AutoMigrate(&dbPrivateKey{})
	if err != nil {
		return nil, err
	}

	return mod, err
}

func init() {
	if err := modules.RegisterModule(keys.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
