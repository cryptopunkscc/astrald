package router

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/log"
	router "github.com/cryptopunkscc/astrald/mod/router/api"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/modules"
)

type Loader struct{}

func (Loader) Load(node modules.Node, assets assets.Store, log *log.Logger) (modules.Module, error) {
	var err error
	var mod = &Module{
		node:   node,
		log:    log.Tag(router.ModuleName),
		routes: make(map[string]id.Identity),
	}

	_ = assets.LoadYAML(router.ModuleName, &mod.config)

	mod.keys, err = assets.KeyStore()

	return mod, err
}

func init() {
	if err := modules.RegisterModule(router.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
