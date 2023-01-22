package keepalive

import "github.com/cryptopunkscc/astrald/node"

const ModuleName = "net.keepalive"

type Loader struct{}

func (Loader) Load(node *node.Node) (node.Module, error) {
	mod := &Module{Node: node}

	return mod, nil
}

func (Loader) Name() string {
	return ModuleName
}
