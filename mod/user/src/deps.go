package user

import (
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/sets"
	"github.com/cryptopunkscc/astrald/mod/user"
)

type Deps struct {
	Admin   admin.Module
	Apphost apphost.Module
	Auth    auth.Module
	Content content.Module
	Dir     dir.Module
	Objects objects.Module
	Keys    keys.Module
	Sets    sets.Module
}

func (mod *Module) LoadDependencies() (err error) {
	err = core.Inject(mod.node, &mod.Deps)
	if err != nil {
		return
	}

	mod.Admin.AddCommand(user.ModuleName, NewAdmin(mod))

	mod.Dir.AddDescriber(mod)
	mod.Objects.AddObject(&user.NodeContract{})
	mod.Objects.AddObject(&user.SignedNodeContract{})

	return
}
