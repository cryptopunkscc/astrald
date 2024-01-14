package fwd

import (
	"github.com/cryptopunkscc/astrald/mod/tcp"
	"github.com/cryptopunkscc/astrald/mod/tor"
	"github.com/cryptopunkscc/astrald/node/modules"
)

func (mod *Module) LoadDependencies() error {
	mod.tcp, _ = modules.Load[tcp.Module](mod.node, tcp.ModuleName)
	mod.tor, _ = modules.Load[tor.Module](mod.node, tor.ModuleName)

	return nil
}
