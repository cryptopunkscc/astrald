package tcpfwd

import (
	"github.com/cryptopunkscc/astrald/node"
)

const ModuleName = "net.tcpfwd"

type Loader struct{}

func (Loader) Load(node *node.Node) (node.Module, error) {
	mod := &Module{node: node}

	if err := node.ConfigStore().LoadYAML("tcpfwd", &mod.config); err != nil {
		log.Errorv(2, "error loading config: %s", err)
	}

	return mod, nil
}

func (Loader) Name() string {
	return ModuleName
}
