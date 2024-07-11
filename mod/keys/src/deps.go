package keys

import (
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/node/modules"
)

func (mod *Module) LoadDependencies() error {
	var err error

	mod.dir, err = modules.Load[dir.Module](mod.node, dir.ModuleName)
	if err != nil {
		return err
	}

	mod.objects, err = modules.Load[objects.Module](mod.node, objects.ModuleName)
	if err != nil {
		return err
	}

	mod.content, err = modules.Load[content.Module](mod.node, content.ModuleName)
	if err != nil {
		return err
	}

	mod.objects.AddDescriber(mod)

	if adm, err := modules.Load[admin.Module](mod.node, admin.ModuleName); err == nil {
		adm.AddCommand(keys.ModuleName, NewAdmin(mod))
	}

	mod.objects.AddPrototypes(keys.KeyDesc{})

	return nil
}
