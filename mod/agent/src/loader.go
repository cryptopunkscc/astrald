package agent

import (
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node"
)

const ModuleName = "agent"

type Loader struct{}

func (Loader) Load(node node.Node, _ assets.Assets, log *log.Logger) (node.Module, error) {
	mod := &Module{node: node, log: log}

	return mod, nil
}

func init() {
	if err := core.RegisterModule(ModuleName, Loader{}); err != nil {
		panic(err)
	}
}
