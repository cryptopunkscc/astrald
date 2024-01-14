package apphost

import (
	"github.com/cryptopunkscc/astrald/mod/data"
	"github.com/cryptopunkscc/astrald/mod/discovery"
	"github.com/cryptopunkscc/astrald/node/modules"
)

func (mod *Module) LoadDependencies() error {
	var err error

	mod.data, _ = modules.Load[data.Module](mod.node, data.ModuleName)

	mod.sdp, err = modules.Load[discovery.Module](mod.node, discovery.ModuleName)
	if err == nil {
		mod.sdp.AddServiceDiscoverer(mod)
	}

	return nil
}
