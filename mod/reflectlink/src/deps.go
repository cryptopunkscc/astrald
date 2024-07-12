package reflectlink

import (
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/discovery"
)

func (mod *Module) LoadDependencies() error {
	mod.sdp, _ = core.Load[discovery.Module](mod.node, discovery.ModuleName)

	return nil
}
