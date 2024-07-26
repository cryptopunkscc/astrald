package gateway

import "github.com/cryptopunkscc/astrald/core"

func (mod *Module) LoadDependencies() (err error) {
	err = core.Inject(mod.node, &mod.Deps)
	if err != nil {
		return
	}

	mod.Exonet.SetDialer("gw", mod.dialer)
	mod.Exonet.SetUnpacker("gw", mod)
	mod.Exonet.SetParser("gw", mod)

	return
}
