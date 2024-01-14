package reflectlink

import (
	"github.com/cryptopunkscc/astrald/mod/discovery"
	"github.com/cryptopunkscc/astrald/node/modules"
)

func (mod *Module) LoadDependencies() error {
	mod.sdp, _ = modules.Load[discovery.Module](mod.node, discovery.ModuleName)

	return nil
}
