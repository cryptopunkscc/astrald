package presence

import (
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/modules"
)

const ModuleName = "presence"

type Loader struct{}

func (Loader) Load(node modules.Node, assets assets.Store, log *log.Logger) (modules.Module, error) {
	mod := &Module{
		node:    node,
		config:  defaultConfig,
		entries: make(map[string]*entry),
		skip:    make(map[string]struct{}),
		log:     log,
	}
	mod.events.SetParent(node.Events())

	_ = assets.LoadYAML("presence", &mod.config)

	return mod, nil
}

func init() {
	if err := modules.RegisterModule(ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
