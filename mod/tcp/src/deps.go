package tcp

import (
	"context"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/tcp"
)

func (mod *Module) LoadDependencies() (err error) {
	mod.nodes, err = core.Load[nodes.Module](mod.node, nodes.ModuleName)
	if err != nil {
		return
	}

	mod.exonet, err = core.Load[exonet.Module](mod.node, exonet.ModuleName)
	if err != nil {
		return
	}

	// inject admin command
	if adm, err := core.Load[admin.Module](mod.node, admin.ModuleName); err == nil {
		adm.AddCommand(tcp.ModuleName, NewAdmin(mod))
	}

	return nil
}

func (mod *Module) Prepare(ctx context.Context) (err error) {
	mod.exonet.SetDialer("tcp", mod)
	mod.exonet.SetParser("tcp", mod)
	mod.exonet.SetUnpacker("tcp", mod)
	mod.exonet.AddResolver(mod)
	return
}
