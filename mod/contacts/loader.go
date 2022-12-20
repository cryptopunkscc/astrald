package contacts

import "github.com/cryptopunkscc/astrald/node"

const ModuleName = "contacts"

type Loader struct{}

func (Loader) Load(node *node.Node) (node.Module, error) {
	mod := &Contacts{node: node}

	return mod, nil
}

func (Loader) Name() string {
	return ModuleName
}
