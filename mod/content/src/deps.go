package content

import (
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/content"
)

func (mod *Module) LoadDependencies() (err error) {
	err = core.Inject(mod.node, &mod.Deps)
	if err != nil {
		return
	}

	mod.Admin.AddCommand(content.ModuleName, NewAdmin(mod))
	mod.Objects.AddDescriber(mod)
	mod.Objects.AddPrototypes(content.LabelDesc{}, content.TypeDesc{})

	mod.setReady()

	return
}
