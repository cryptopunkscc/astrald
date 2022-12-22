package apphost

import (
	"github.com/cryptopunkscc/astrald/node"
	"net"
)

const ModuleName = "apphost"

type Loader struct{}

func (Loader) Load(node *node.Node) (node.Module, error) {
	mod := &Module{
		node:        node,
		listeners:   make([]net.Listener, 0),
		clientConns: make(chan net.Conn),
	}

	return mod, nil
}

func (Loader) Name() string {
	return ModuleName
}
