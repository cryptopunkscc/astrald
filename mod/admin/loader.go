package admin

import (
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/modules"
)

const ModuleName = "admin"

type Loader struct{}

func (Loader) Load(node modules.Node, assets assets.Store) (modules.Module, error) {
	mod := &Module{
		config:   defaultConfig,
		node:     node,
		commands: make(map[string]Command),
		log:      log.Tag(ModuleName),
	}

	assets.LoadYAML(ModuleName, &mod.config)

	mod.AddCommand("help", &CmdHelp{mod: mod})
	mod.AddCommand("tracker", &CmdTracker{mod: mod})
	mod.AddCommand("net", &CmdNet{mod: mod})
	mod.AddCommand("enter", &CmdEnter{mod: mod})

	return mod, nil
}

func init() {
	if err := modules.RegisterModule(ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
