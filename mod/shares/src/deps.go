package shares

import (
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/sets"
	"github.com/cryptopunkscc/astrald/mod/shares"
)

func (mod *Module) LoadDependencies() error {
	var err error

	mod.objects, err = core.Load[objects.Module](mod.node, objects.ModuleName)
	if err != nil {
		return err
	}

	mod.sets, err = core.Load[sets.Module](mod.node, sets.ModuleName)
	if err != nil {
		return err
	}

	if adm, err := core.Load[admin.Module](mod.node, admin.ModuleName); err == nil {
		adm.AddCommand(shares.ModuleName, NewAdmin(mod))
	}

	mod.objects.AddOpener(mod, 10)
	mod.objects.AddDescriber(mod)

	return err
}
