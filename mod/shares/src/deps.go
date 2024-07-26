package shares

import (
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/shares"
)

func (mod *Module) LoadDependencies() (err error) {
	err = core.Inject(mod.node, &mod.Deps)
	if err != nil {
		return
	}

	mod.Admin.AddCommand(shares.ModuleName, NewAdmin(mod))

	mod.Auth.AddAuthorizer(&Authorizer{mod: mod})

	mod.Objects.AddOpener(mod, 10)
	mod.Objects.AddDescriber(mod)

	return
}
