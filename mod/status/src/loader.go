package status

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/log"
)

const ModuleName = "status"

type Loader struct{}

func (Loader) Load(node astral.Node, assets assets.Assets, log *log.Logger) (core.Module, error) {
	mod := &Module{
		node:       node,
		config:     defaultConfig,
		log:        log,
		setVisible: make(chan bool, 1),
	}
	mod.Provider = NewProvider(mod)

	_ = assets.LoadYAML(ModuleName, &mod.config)

	return mod, nil
}

func init() {
	if err := core.RegisterModule(ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
