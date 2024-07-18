package apphost

import (
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/discovery"
)

func (mod *Module) LoadDependencies() error {
	var err error

	mod.content, _ = core.Load[content.Module](mod.node, content.ModuleName)

	mod.sdp, err = core.Load[discovery.Module](mod.node, discovery.ModuleName)
	if err == nil {
		mod.sdp.AddServiceDiscoverer(mod)
	}

	mod.dir, err = core.Load[dir.Module](mod.node, dir.ModuleName)
	if err != nil {
		return err
	}

	return nil
}
