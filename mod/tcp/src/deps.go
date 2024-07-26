package tcp

import (
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/tcp"
)

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

	return
}
