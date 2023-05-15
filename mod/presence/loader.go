package presence

import (
	_log "github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node/config"
	"github.com/cryptopunkscc/astrald/node/modules"
)

const ModuleName = "presence"

var log = _log.Tag(ModuleName)

type Loader struct{}

func (Loader) Load(node modules.Node, configStore config.Store) (modules.Module, error) {
	mod := &Module{
		node:    node,
		config:  defaultConfig,
		entries: make(map[string]*entry),
		skip:    make(map[string]struct{}),
	}
	mod.events.SetParent(node.Events())

	if err := configStore.LoadYAML("presence", &mod.config); err != nil {
		log.Errorv(2, "error loading config: %s", err)
	}

	return mod, nil
}

func (Loader) Name() string {
	return ModuleName
}

func init() {
	if err := modules.RegisterModule(ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
