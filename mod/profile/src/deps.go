package profile

import (
	"github.com/cryptopunkscc/astrald/mod/discovery"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/node/modules"
)

func (mod *Module) LoadDependencies() (err error) {
	mod.sdp, _ = modules.Load[discovery.Module](mod.node, discovery.ModuleName)

	mod.nodes, err = modules.Load[nodes.Module](mod.node, nodes.ModuleName)

	return
}
