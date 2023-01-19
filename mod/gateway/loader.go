package gateway

import "github.com/cryptopunkscc/astrald/node"

const ModuleName = "gateway"

type Loader struct{}

func (Loader) Load(node *node.Node) (node.Module, error) {
	mod := &Gateway{node: node}

	return mod, nil
}

func (Loader) Name() string {
	return ModuleName
}