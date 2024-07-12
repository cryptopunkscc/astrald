package user

import (
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/user"
	"github.com/cryptopunkscc/astrald/node"
)

type Loader struct{}

func (Loader) Load(node node.Node, assets assets.Assets, log *log.Logger) (node.Module, error) {
	var err error
	var mod = &Module{
		node:   node,
		config: defaultConfig,
		routes: core.NewPrefixRouter(false),
		log:    log,
		assets: assets,
	}

	mod.routes.EnableParams = true
	mod.profileService = &ProfileService{Module: mod}

	err = assets.LoadYAML(user.ModuleName, &mod.config)
	if err != nil {

	}

	mod.db = assets.Database()

	err = mod.db.AutoMigrate(&dbIdentity{}, &dbNodeContract{})
	if err != nil {
		return nil, err
	}

	mod.node.Auth().Add(&Authorizer{mod: mod})

	return mod, err
}

func init() {
	if err := core.RegisterModule(user.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
