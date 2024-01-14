package admin

import (
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/modules"
)

type Loader struct{}

func (Loader) Load(node modules.Node, assets assets.Store, log *log.Logger) (modules.Module, error) {
	mod := &Module{
		config:   defaultConfig,
		node:     node,
		assets:   assets,
		commands: make(map[string]admin.Command),
		log:      log,
	}

	_ = assets.LoadYAML(admin.ModuleName, &mod.config)

	_ = mod.AddCommand("help", &CmdHelp{mod: mod})
	_ = mod.AddCommand("tracker", NewCmdTracker(mod))
	_ = mod.AddCommand("net", &CmdNet{mod: mod})
	_ = mod.AddCommand("use", &CmdUse{mod: mod})
	_ = mod.AddCommand("node", &CmdNode{mod: mod})
	_ = mod.AddCommand(admin.ModuleName, NewCmdAdmin(mod))

	return mod, nil
}

func init() {
	if err := modules.RegisterModule(admin.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
