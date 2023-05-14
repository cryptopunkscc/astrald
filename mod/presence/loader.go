package presence

import (
	_log "github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/config"
)

const ModuleName = "presence"

var log = _log.Tag(ModuleName)

type Loader struct{}

func (Loader) Load(node node.Node, configStore config.Store) (node.Module, error) {
	mod := &Module{
		node:    node,
		config:  defaultConfig,
		entries: make(map[string]*entry),
		skip:    make(map[string]struct{}),
	}
	mod.events.SetParent(node.Events())

	if err := node.ConfigStore().LoadYAML("presence", &mod.config); err != nil {
		log.Errorv(2, "error loading config: %s", err)
	}

	return mod, nil
}

func (Loader) Name() string {
	return ModuleName
}
