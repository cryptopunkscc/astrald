package utp

import (
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

// Deps represents the dependencies required by the src UDP module.
type Deps struct {
	Exonet  exonet.Module
	Nodes   nodes.Module
	Objects objects.Module
}

func (mod *Module) LoadDependencies() (err error) {
	err = core.Inject(mod.node, &mod.Deps)
	if err != nil {
		return
	}

	mod.Exonet.SetDialer("utp", mod)
	mod.Exonet.SetParser("utp", mod)
	mod.Exonet.SetUnpacker("utp", mod)
	mod.Nodes.AddResolver(mod)

	return
}
