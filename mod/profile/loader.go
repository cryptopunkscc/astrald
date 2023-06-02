package profile

import (
	_log "github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/modules"
)

const ModuleName = "profile"

type Loader struct{}

func (Loader) Load(node modules.Node, _ assets.Store) (modules.Module, error) {
	mod := &Module{
		node: node,
		log:  _log.Tag(ModuleName),
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
