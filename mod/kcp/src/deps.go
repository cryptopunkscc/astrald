package kcp

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/ip"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

type Deps struct {
	Exonet  exonet.Module
	Nodes   nodes.Module
	Objects objects.Module
	IP      ip.Module
}

func (mod *Module) LoadDependencies(*astral.Context) (err error) {
	err = core.Inject(mod.node, &mod.Deps)
	if err != nil {
		return
	}

	mod.Exonet.SetDialer("kcp", mod)
	mod.Exonet.SetParser("kcp", mod)
	mod.Exonet.SetUnpacker("kcp", mod)
	mod.Nodes.AddResolver(mod)

	return
}
