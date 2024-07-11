package agent

import (
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/node/modules"
)

func (mod *Module) LoadDependencies() (err error) {
	mod.dir, err = modules.Load[dir.Module](mod.node, dir.ModuleName)
	if err != nil {
		return err
	}

	return nil
}
