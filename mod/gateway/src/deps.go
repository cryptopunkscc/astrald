package gateway

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core"
)

func (mod *Module) LoadDependencies(*astral.Context) (err error) {
	err = core.Inject(mod.node, &mod.Deps)
	if err != nil {
		return
	}

	mod.Exonet.SetDialer("gw", mod)
	mod.Exonet.SetUnpacker("gw", mod)
	mod.Exonet.SetParser("gw", mod)
	mod.ops.AddStructPrefix(mod, "Op")
	mod.Services.AddDiscoverer(mod)

	return
}
