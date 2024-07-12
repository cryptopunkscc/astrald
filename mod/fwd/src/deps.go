package fwd

import (
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/tcp"
	"github.com/cryptopunkscc/astrald/mod/tor"
)

func (mod *Module) LoadDependencies() error {
	mod.tcp, _ = core.Load[tcp.Module](mod.node, tcp.ModuleName)
	mod.tor, _ = core.Load[tor.Module](mod.node, tor.ModuleName)

	return nil
}
