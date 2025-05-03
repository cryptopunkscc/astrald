package tor

import (
	"github.com/cryptopunkscc/astrald/core"
)

func (mod *Module) LoadDependencies() (err error) {
	err = core.Inject(mod.node, &mod.Deps)
	if err != nil {
		return
	}

	mod.Exonet.SetDialer("tor", mod)
	mod.Exonet.SetParser("tor", mod)
	mod.Exonet.SetUnpacker("tor", mod)
	mod.Nodes.AddResolver(mod)

	return nil
}
