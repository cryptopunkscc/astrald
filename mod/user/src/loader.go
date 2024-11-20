package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/user"
)

type Loader struct{}

func (Loader) Load(node astral.Node, assets assets.Assets, log *log.Logger) (core.Module, error) {
	var err error
	var mod = &Module{
		node:   node,
		config: defaultConfig,
		log:    log,
		assets: assets,
	}

	mod.provider = NewProvider(mod)

	err = assets.LoadYAML(user.ModuleName, &mod.config)
	if err != nil {

	}

	mod.db = assets.Database()

	err = mod.db.AutoMigrate(&dbIdentity{}, &dbNodeContract{}, &dbContact{})
	if err != nil {
		return nil, err
	}

	return mod, err
}

func init() {
	if err := core.RegisterModule(user.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
