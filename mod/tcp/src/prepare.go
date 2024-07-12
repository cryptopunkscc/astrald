package tcp

import (
	"context"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/tcp"
)

func (mod *Module) Prepare(ctx context.Context) (err error) {
	mod.nodes, err = core.Load[nodes.Module](mod.node, nodes.ModuleName)
	if err != nil {
		return
	}

	// inject admin command
	if adm, err := core.Load[admin.Module](mod.node, admin.ModuleName); err == nil {
		adm.AddCommand(tcp.ModuleName, NewAdmin(mod))
	}

	return
}
