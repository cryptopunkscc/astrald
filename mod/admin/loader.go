package admin

import (
	"github.com/cryptopunkscc/astrald/node/config"
	"github.com/cryptopunkscc/astrald/node/modules"
)

const ModuleName = "admin"

type Loader struct{}

func (Loader) Load(node modules.Node, _ config.Store) (modules.Module, error) {
	mod := &Admin{node: node}

	return mod, nil
}

func (Loader) Name() string {
	return ModuleName
}

func init() {
	if err := modules.RegisterModule(ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
