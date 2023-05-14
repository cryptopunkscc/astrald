package apphost

import (
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/config"
	"net"
)

const ModuleName = "apphost"

type Loader struct{}

func (Loader) Load(node node.Node, configStore config.Store) (node.Module, error) {
	mod := &Module{
		config:    defaultConfig,
		node:      node,
		listeners: make([]net.Listener, 0),
		tokens:    make(map[string]string, 0),
	}

	configStore.LoadYAML(ModuleName, &mod.config)

	return mod, nil
}

func (Loader) Name() string {
	return ModuleName
}
