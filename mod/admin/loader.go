package admin

import (
	"github.com/cryptopunkscc/astrald/log"
	. "github.com/cryptopunkscc/astrald/mod/admin/api"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/modules"
)

const ModuleName = "admin"

type Loader struct{}

func (Loader) Load(node modules.Node, assets assets.Store, log *log.Logger) (modules.Module, error) {
	mod := &Module{
		config:   defaultConfig,
		node:     node,
		assets:   assets,
		commands: make(map[string]Command),
		log:      log,
	}

	_ = assets.LoadYAML(ModuleName, &mod.config)

	_ = mod.AddCommand("help", &CmdHelp{mod: mod})
	_ = mod.AddCommand("tracker", NewCmdTracker(mod))
	_ = mod.AddCommand("net", &CmdNet{mod: mod})
	_ = mod.AddCommand("use", &CmdUse{mod: mod})
	_ = mod.AddCommand("node", &CmdNode{mod: mod})

	return mod, nil
}

func init() {
	if err := modules.RegisterModule(ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
