package policy

import (
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/relay"
)

func (mod *Module) LoadDependencies() (err error) {
	mod.nodes, err = core.Load[nodes.Module](mod.node, nodes.ModuleName)
	if err != nil {
		return
	}

	mod.relay, _ = core.Load[relay.Module](mod.node, relay.ModuleName)

	return
}
