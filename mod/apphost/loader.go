package apphost

import (
	"github.com/cryptopunkscc/astrald/node/config"
	"github.com/cryptopunkscc/astrald/node/modules"
	"net"
)

const ModuleName = "apphost"

type Loader struct{}

func (Loader) Load(node modules.Node, configStore config.Store) (modules.Module, error) {
	mod := &Module{
		config:    defaultConfig,
		node:      node,
		listeners: make([]net.Listener, 0),
		tokens:    make(map[string]string, 0),
	}

	configStore.LoadYAML(ModuleName, &mod.config)

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
