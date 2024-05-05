package archives

import (
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/archives"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/shares"
	"github.com/cryptopunkscc/astrald/node/modules"
)

func (mod *Module) LoadDependencies() error {
	var err error

	mod.content, err = modules.Load[content.Module](mod.node, content.ModuleName)
	if err != nil {
		return err
	}

	mod.objects, err = modules.Load[objects.Module](mod.node, objects.ModuleName)
	if err != nil {
		return err
	}

	mod.shares, err = modules.Load[shares.Module](mod.node, shares.ModuleName)
	if err != nil {
		return err
	}

	// inject admin command
	if adm, err := modules.Load[admin.Module](mod.node, admin.ModuleName); err == nil {
		adm.AddCommand(archives.ModuleName, NewAdmin(mod))
	}

	mod.objects.AddPrototypes(archives.ArchiveDesc{}, archives.EntryDesc{})
	mod.objects.AddOpener(mod, 20)
	mod.objects.AddDescriber(mod)
	mod.objects.AddFinder(mod)

	return nil
}
