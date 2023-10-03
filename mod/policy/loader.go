package policy

import (
	_log "github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/modules"
)

const ModuleName = "policy"

type Loader struct{}

func (Loader) Load(node modules.Node, assets assets.Store, log *_log.Logger) (modules.Module, error) {
	mod := &Module{
		node:     node,
		log:      log,
		config:   defaultConfig,
		policies: make(map[*RunningPolicy]struct{}),
	}

	_ = assets.LoadYAML(ModuleName, &mod.config)

	return mod, nil
}

func init() {
	if err := modules.RegisterModule(ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
