package tcp

import (
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/tcp"
)

type Deps struct {
	Admin   admin.Module
	Exonet  exonet.Module
	Nodes   nodes.Module
	Objects objects.Module
}

func (mod *Module) LoadDependencies() (err error) {
	err = core.Inject(mod.node, &mod.Deps)
	if err != nil {
		return
	}

	mod.Admin.AddCommand(tcp.ModuleName, NewAdmin(mod))

	mod.Exonet.SetDialer("tcp", mod)
	mod.Exonet.SetParser("tcp", mod)
	mod.Exonet.SetUnpacker("tcp", mod)
	mod.Exonet.AddResolver(mod)

	mod.Objects.AddObject(&tcp.IP{})
	mod.Objects.AddObject(&Endpoint{})

	return
}
