package policy

import (
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/relay"
	"github.com/cryptopunkscc/astrald/node/modules"
)

func (mod *Module) LoadDependencies() (err error) {
	mod.nodes, err = modules.Load[nodes.Module](mod.node, nodes.ModuleName)
	if err != nil {
		return
	}

	mod.relay, _ = modules.Load[relay.Module](mod.node, relay.ModuleName)

	return
}
