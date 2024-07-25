package apphost

import (
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/dir"
)

func (mod *Module) LoadDependencies() (err error) {
	mod.content, _ = core.Load[content.Module](mod.node, content.ModuleName)

	mod.dir, err = core.Load[dir.Module](mod.node, dir.ModuleName)

	return
}
