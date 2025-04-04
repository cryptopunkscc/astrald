package admin

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/mod/admin"
)

type Loader struct{}

func (Loader) Load(node astral.Node, assets assets.Assets, log *log.Logger) (core.Module, error) {
	mod := &Module{
		config:   defaultConfig,
		node:     node,
		assets:   assets,
		commands: make(map[string]admin.Command),
		log:      log,
	}

	_ = assets.LoadYAML(admin.ModuleName, &mod.config)

	_ = mod.AddCommand("help", &CmdHelp{mod: mod})
	_ = mod.AddCommand("use", &CmdUse{mod: mod})
	_ = mod.AddCommand("sudo", &CmdSudo{mod: mod})
	_ = mod.AddCommand("node", &CmdNode{mod: mod})
	_ = mod.AddCommand(admin.ModuleName, NewCmdAdmin(mod))

	return mod, nil
}

func init() {
	if err := core.RegisterModule(admin.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
