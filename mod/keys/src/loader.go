package keys

import (
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/modules"
)

type Loader struct{}

func (Loader) Load(node modules.Node, assets assets.Assets, log *log.Logger) (modules.Module, error) {
	var err error
	var mod = &Module{
		node:   node,
		log:    log,
		assets: assets,
	}

	_ = assets.LoadYAML(keys.ModuleName, &mod.config)

	mod.db = mod.assets.Database()

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
