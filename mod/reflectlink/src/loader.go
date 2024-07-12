package reflectlink

import (
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	_log "github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node"
)

const ModuleName = "net.reflectlink"

type Loader struct{}

func (Loader) Load(node node.Node, _ assets.Assets, log *_log.Logger) (node.Module, error) {
	mod := &Module{
		node: node,
		log:  log,
	}

	return mod, nil
}

func init() {
	if err := core.RegisterModule(ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
