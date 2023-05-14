package tcpfwd

import (
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/config"
)

const ModuleName = "net.tcpfwd"

type Loader struct{}

func (Loader) Load(node node.Node, configStore config.Store) (node.Module, error) {
	mod := &Module{
		node:   node,
		config: defaultConfig,
	}

	configStore.LoadYAML("tcpfwd", &mod.config)

	return mod, nil
}

func (Loader) Name() string {
	return ModuleName
}
