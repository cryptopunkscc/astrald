package apphost

import (
	"errors"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/config"
	"net"
)

const ModuleName = "apphost"

type Loader struct{}

func (Loader) Load(node node.Node) (node.Module, error) {
	mod := &Module{
		config:    defaultConfig,
		node:      node,
		listeners: make([]net.Listener, 0),
		tokens:    make(map[string]string, 0),
	}

	err := node.ConfigStore().LoadYAML(ModuleName, &mod.config)

	switch {
	case err == nil:

	case errors.Is(err, config.ErrNotFound):
		log.Logv(2, "config not found")

	default:
		log.Errorv(1, "error loading config: %s", err)
	}

	return mod, nil
}

func (Loader) Name() string {
	return ModuleName
}
