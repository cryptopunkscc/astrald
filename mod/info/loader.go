package info

import "github.com/cryptopunkscc/astrald/node"

const ModuleName = "info"

type Loader struct{}

func (Loader) Load(node *node.Node) (node.Module, error) {
	mod := &Info{node: node}

	return mod, nil
}

func (Loader) Name() string {
	return ModuleName
}
