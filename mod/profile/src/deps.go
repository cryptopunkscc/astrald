package profile

import (
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

func (mod *Module) LoadDependencies() (err error) {
	mod.nodes, err = core.Load[nodes.Module](mod.node, nodes.ModuleName)
	if err != nil {
		return err
	}

	mod.exonet, err = core.Load[exonet.Module](mod.node, exonet.ModuleName)
	if err != nil {
		return err
	}

	mod.dir, err = core.Load[dir.Module](mod.node, dir.ModuleName)
	if err != nil {
		return err
	}

	mod.objects, err = core.Load[objects.Module](mod.node, objects.ModuleName)
	if err != nil {
		return err
	}

	mod.objects.AddReceiver(mod)

	return
}
