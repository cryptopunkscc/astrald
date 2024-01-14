package user

import (
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/user"
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

	mod.profileHandler = &ProfileHandler{Module: mod}

	mod.db, err = assets.OpenDB(user.ModuleName)
	if err != nil {
		return nil, err
	}

	err = mod.db.AutoMigrate(&dbIdentity{})
	if err != nil {
		return nil, err
	}

	_ = assets.LoadYAML(user.ModuleName, &mod.config)

	return mod, err
}

func init() {
	if err := modules.RegisterModule(user.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
