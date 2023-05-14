package nat

import (
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/config"
)

const ModuleName = "nat-tcp"

type Loader struct{}

func (Loader) Load(node node.Node, configStore config.Store) (node.Module, error) {
	mod := &Module{node: node}

	return mod, nil
}

func (Loader) Name() string {
	return ModuleName
}
