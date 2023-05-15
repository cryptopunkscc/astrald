package tcpfwd

import (
	"github.com/cryptopunkscc/astrald/node/config"
	"github.com/cryptopunkscc/astrald/node/modules"
)

const ModuleName = "net.tcpfwd"

type Loader struct{}

func (Loader) Load(node modules.Node, configStore config.Store) (modules.Module, error) {
	mod := &Module{
		node:   node,
		config: defaultConfig,
	}

	configStore.LoadYAML("tcpfwd", &mod.config)

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
