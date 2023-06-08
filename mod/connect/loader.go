package connect

import (
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/modules"
)

const ModuleName = "connect"

type Loader struct{}

func (Loader) Load(node modules.Node, _ assets.Store, _ *log.Logger) (modules.Module, error) {
	mod := &Connect{node: node}

	return mod, nil
}

func init() {
	if err := modules.RegisterModule(ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
