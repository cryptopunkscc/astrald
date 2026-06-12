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

// LoadDependencies injects required modules and registers the object-read
// authorizer as a side effect.
func (mod *Module) LoadDependencies(*astral.Context) (err error) {
	if err = core.Inject(mod.node, &mod.Deps); err != nil {
		return
	}
	mod.Auth.Add(auth.Func[*objects.ReadObjectAction](mod.AuthorizeObjectsRead))
	return
}
