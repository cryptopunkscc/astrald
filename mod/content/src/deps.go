package content

import (
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/fs"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

type Deps struct {
	Admin   admin.Module
	FS      fs.Module
	Objects objects.Module
}

func (mod *Module) LoadDependencies() (err error) {
	err = core.Inject(mod.node, &mod.Deps)
	if err != nil {
		return
	}

	mod.Admin.AddCommand(content.ModuleName, NewAdmin(mod))
	mod.Objects.AddDescriber(mod)
	mod.Objects.AddObject(&content.ObjectDescriptor{})

	mod.setReady()

	return
}
