package utp

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/ip"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

// Deps represents the dependencies required by the uTP module.
type Deps struct {
	Exonet  exonet.Module
	Nodes   nodes.Module
	Objects objects.Module
	IP      ip.Module
}

// LoadDependencies injects required modules and registers this module as the
// uTP dialer, parser, unpacker, and endpoint resolver with the exonet and nodes subsystems.
func (mod *Module) LoadDependencies(*astral.Context) (err error) {
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
