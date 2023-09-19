package sdp

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/modules"
)

const ModuleName = "sdp"

type Loader struct{}

func (Loader) Load(node modules.Node, assets assets.Store, log *log.Logger) (modules.Module, error) {
	var err error
	var mod = &Module{
		node:    node,
		config:  defaultConfig,
		log:     log,
		assets:  assets,
		sources: map[Source]struct{}{},
		cache:   map[string][]ServiceEntry{},
		routes:  map[string]id.Identity{},
	}

	mod.events.SetParent(node.Events())
	mod.AddSource(NewLocalServices(node.Services()))

	_ = assets.LoadYAML(ModuleName, &mod.config)

	return mod, err
}

func init() {
	if err := modules.RegisterModule(ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
