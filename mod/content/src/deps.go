package content

import (
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/fs"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/node/modules"
)

func (mod *Module) LoadDependencies() error {
	var err error

	// required
	mod.objects, err = modules.Load[objects.Module](mod.node, objects.ModuleName)
	if err != nil {
		return err
	}

	mod.objects.AddDescriber(mod)
	mod.objects.AddPrototypes(content.LabelDesc{}, content.TypeDesc{})

	// optional
	mod.fs, _ = modules.Load[fs.Module](mod.node, fs.ModuleName)

	// inject admin command
	if adm, err := modules.Load[admin.Module](mod.node, admin.ModuleName); err == nil {
		adm.AddCommand(content.ModuleName, NewAdmin(mod))
	}

	mod.setReady()

	return nil
}
