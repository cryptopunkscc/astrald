package archives

import (
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/archives"
)

func (mod *Module) LoadDependencies() (err error) {
	err = core.Inject(mod.node, &mod.Deps)
	if err != nil {
		return
	}

	mod.Admin.AddCommand(archives.ModuleName, NewAdmin(mod))

	mod.Auth.AddAuthorizer(&Authorizer{mod: mod})

	mod.Objects.AddPrototypes(archives.ArchiveDesc{}, archives.EntryDesc{})
	mod.Objects.AddOpener(mod, 20)
	mod.Objects.AddDescriber(mod)
	mod.Objects.AddSearcher(mod)

	return nil
}
