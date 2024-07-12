package keys

import (
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

func (mod *Module) LoadDependencies() error {
	var err error

	mod.dir, err = core.Load[dir.Module](mod.node, dir.ModuleName)
	if err != nil {
		return err
	}

	mod.objects, err = core.Load[objects.Module](mod.node, objects.ModuleName)
	if err != nil {
		return err
	}

	mod.content, err = core.Load[content.Module](mod.node, content.ModuleName)
	if err != nil {
		return err
	}

	mod.objects.AddDescriber(mod)

	if adm, err := core.Load[admin.Module](mod.node, admin.ModuleName); err == nil {
		adm.AddCommand(keys.ModuleName, NewAdmin(mod))
	}

	mod.objects.AddPrototypes(keys.KeyDesc{})

	return nil
}
