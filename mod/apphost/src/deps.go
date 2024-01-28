package apphost

import (
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/discovery"
	"github.com/cryptopunkscc/astrald/node/modules"
)

func (mod *Module) LoadDependencies() error {
	var err error

	mod.content, _ = modules.Load[content.Module](mod.node, content.ModuleName)

	mod.sdp, err = modules.Load[discovery.Module](mod.node, discovery.ModuleName)
	if err == nil {
		mod.sdp.AddServiceDiscoverer(mod)
	}

	return nil
}
