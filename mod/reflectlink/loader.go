package reflectlink

import (
	"github.com/cryptopunkscc/astrald/node"
)

const ModuleName = "net.reflectlink"

type Loader struct{}

func (Loader) Load(node node.Node) (node.Module, error) {
	mod := &Module{node: node}

	return mod, nil
}

func (Loader) Name() string {
	return ModuleName
}
