package reflectlink

import (
	_log "github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/modules"
)

const ModuleName = "net.reflectlink"

type Loader struct{}

func (Loader) Load(node modules.Node, _ assets.Store, log *_log.Logger) (modules.Module, error) {
	mod := &Module{
		node: node,
		log:  log,
	}

	return mod, nil
}

func init() {
	if err := modules.RegisterModule(ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
