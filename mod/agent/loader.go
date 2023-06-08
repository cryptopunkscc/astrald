package agent

import (
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/modules"
)

const ModuleName = "agent"

type Loader struct{}

func (Loader) Load(node modules.Node, _ assets.Store, log *log.Logger) (modules.Module, error) {
	mod := &Module{node: node, log: log}

	return mod, nil
}

func init() {
	if err := modules.RegisterModule(ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
