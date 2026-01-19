package crypto

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/mod/crypto"
)

type Loader struct{}

func (Loader) Load(node astral.Node, assets assets.Assets, log *log.Logger) (core.Module, error) {
	var err error
	var mod = &Module{
		config: defaultConfig,
		node:   node,
		log:    log,
		assets: assets,
	}

	_ = assets.LoadYAML(crypto.ModuleName, &mod.config)

	mod.scope.AddStruct(mod, "Op")

	mod.db, err = newDB(mod.assets.Database())
	if err != nil {
		return nil, err
	}
	
	return mod, err
}

func init() {
	if err := core.RegisterModule(crypto.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
