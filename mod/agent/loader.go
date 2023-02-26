package agent

import (
	"github.com/cryptopunkscc/astrald/node"
)

const ModuleName = "agent"

type Loader struct{}

func (Loader) Load(node *node.Node) (node.Module, error) {
	mod := &Module{node: node}

	return mod, nil
}

func (Loader) Name() string {
	return ModuleName
}
