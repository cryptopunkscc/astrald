package discovery

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/modules"
)

const ModuleName = "discovery"

type Loader struct{}

func (Loader) Load(node modules.Node, assets assets.Store, log *log.Logger) (modules.Module, error) {
	var err error
	var mod = &Module{
		node:    node,
		config:  defaultConfig,
		log:     log,
		sources: map[Source]id.Identity{},
		cache:   map[string][]ServiceEntry{},
	}

	mod.events.SetParent(node.Events())

	_ = assets.LoadYAML(ModuleName, &mod.config)

	return mod, err
}

func init() {
	if err := modules.RegisterModule(ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
