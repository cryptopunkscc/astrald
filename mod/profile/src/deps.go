package profile

import (
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/discovery"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nodes"
)

func (mod *Module) LoadDependencies() (err error) {
	mod.sdp, _ = core.Load[discovery.Module](mod.node, discovery.ModuleName)

	mod.nodes, err = core.Load[nodes.Module](mod.node, nodes.ModuleName)
	if err != nil {
		return err
	}

	mod.exonet, err = core.Load[exonet.Module](mod.node, exonet.ModuleName)

	return
}
