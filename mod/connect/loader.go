package connect

import "github.com/cryptopunkscc/astrald/node"

const ModuleName = "connect"

type Loader struct{}

func (Loader) Load(node *node.Node) (node.Module, error) {
	mod := &Connect{node: node}

	return mod, nil
}

func (Loader) Name() string {
	return ModuleName
}
