package shell

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type Loader struct{}

func (Loader) Load(node astral.Node, assets assets.Assets, log *log.Logger) (core.Module, error) {
	var err error
	var mod = &Module{
		node:   node,
		config: defaultConfig,
		log:    log,
		assets: assets,
	}
	mod.root.Log = log

	_ = assets.LoadYAML(shell.ModuleName, &mod.config)

	err = mod.root.AddStruct(mod, "Op")

	return mod, err
}

func init() {
	if err := core.RegisterModule(shell.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
