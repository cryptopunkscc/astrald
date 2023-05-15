package admin

import (
	"github.com/cryptopunkscc/astrald/node/config"
	"github.com/cryptopunkscc/astrald/node/modules"
)

const ModuleName = "admin"

type Loader struct{}

func (Loader) Load(node modules.Node, configStore config.Store) (modules.Module, error) {
	mod := &Module{
		config:   defaultConfig,
		node:     node,
		commands: make(map[string]Command),
	}

	configStore.LoadYAML(ModuleName, &mod.config)

	mod.AddCommand("help", &HelpCommand{mod: mod})
	mod.AddCommand("tracker", &TrackerCommand{node: node})
	mod.AddCommand("net", &NetCommand{mod: mod})

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
