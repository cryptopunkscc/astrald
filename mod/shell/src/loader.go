package shell

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/lib/ops"
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
	mod.root.OnError = func(err error, query *astral.Query) {
		mod.log.Logv(1, "[%v] error processing '%v': %v", query.Nonce, query.Query, err)
	}

	_ = assets.LoadYAML(shell.ModuleName, &mod.config)

	var scope ops.Set
	err = scope.AddStruct(mod, "Op")
	if err != nil {
		return nil, err
	}

	err = mod.root.AddSet(shell.ModuleName, &scope)

	return mod, err
}

func init() {
	if err := core.RegisterModule(shell.ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
