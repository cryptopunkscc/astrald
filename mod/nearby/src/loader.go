package nearby

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
)

const ModuleName = "nearby"

type Loader struct{}

func (Loader) Load(node astral.Node, assets assets.Assets, log *log.Logger) (core.Module, error) {
	mod := &Module{
		node:       node,
		config:     defaultConfig,
		log:        log,
		setVisible: make(chan bool, 1),
	}

	_ = assets.LoadYAML(ModuleName, &mod.config)

	err := mod.scope.AddStructPrefix(mod, "Op")
	if err != nil {
		return nil, err
	}

	return mod, nil
}

func init() {
	if err := core.RegisterModule(ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
