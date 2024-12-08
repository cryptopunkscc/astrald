package tor

import (
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/tor"
)

func (mod *Module) LoadDependencies() (err error) {
	err = core.Inject(mod.node, &mod.Deps)
	if err != nil {
		return
	}

	mod.Admin.AddCommand(tor.ModuleName, NewAdmin(mod))

	mod.Exonet.SetDialer("tor", mod)
	mod.Exonet.SetParser("tor", mod)
	mod.Exonet.SetUnpacker("tor", mod)
	mod.Exonet.AddResolver(mod)

	return nil
}
