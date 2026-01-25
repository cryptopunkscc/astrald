package tcp

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	ipmod "github.com/cryptopunkscc/astrald/mod/ip"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

type Deps struct {
	Exonet  exonet.Module
	Nodes   nodes.Module
	Objects objects.Module
	IP      ipmod.Module
}

func (mod *Module) LoadDependencies(*astral.Context) (err error) {
	err = core.Inject(mod.node, &mod.Deps)
	if err != nil {
		return
	}

	mod.Exonet.SetDialer("tcp", mod)
	mod.Exonet.SetParser("tcp", mod)
	mod.Exonet.SetUnpacker("tcp", mod)
	mod.Nodes.AddResolver(mod)

	return
}
