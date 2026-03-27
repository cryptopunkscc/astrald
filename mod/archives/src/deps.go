package archives

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

type Deps struct {
	Auth    auth.Module
	Objects objects.Module
}

func (mod *Module) LoadDependencies(*astral.Context) (err error) {
	if err = core.Inject(mod.node, &mod.Deps); err != nil {
		return
	}
	mod.Auth.AddAuthorizer(objects.ActionRead, auth.Func[*astral.ObjectID](mod.AuthorizeObjectsRead))
	return
}
