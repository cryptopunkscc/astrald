package user

import (
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/user"
)

func (mod *Module) LoadDependencies() (err error) {
	err = core.Inject(mod.node, &mod.Deps)
	if err != nil {
		return
	}

	mod.Admin.AddCommand(user.ModuleName, NewAdmin(mod))

	mod.Auth.AddAuthorizer(&Authorizer{mod: mod})
	mod.Dir.AddDescriber(mod)
	mod.Objects.AddObject(&user.NodeContract{})

	return
}
