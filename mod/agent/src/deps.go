package agent

import (
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/dir"
)

func (mod *Module) LoadDependencies() (err error) {
	mod.dir, err = core.Load[dir.Module](mod.node, dir.ModuleName)
	if err != nil {
		return err
	}

	return nil
}
