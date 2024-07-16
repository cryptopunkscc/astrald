package reflectlink

import (
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/discovery"
	"github.com/cryptopunkscc/astrald/mod/exonet"
)

func (mod *Module) LoadDependencies() error {
	mod.sdp, _ = core.Load[discovery.Module](mod.node, discovery.ModuleName)

	mod.exonet, _ = core.Load[exonet.Module](mod.node, exonet.ModuleName)

	return nil
}
