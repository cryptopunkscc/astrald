package media

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/media"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type Deps struct {
	Auth    auth.Module
	Objects objects.Module
	Shell   shell.Module
}

func (mod *Module) LoadDependencies(*astral.Context) (err error) {
	if err = core.Inject(mod.node, &mod.Deps); err != nil {
		return
	}
	mod.Auth.Add(objects.ActionRead, auth.Func[*media.AudioFile](mod.AuthorizeObjectsReadByFile))
	mod.Auth.Add(objects.ActionRead, auth.Func[*astral.ObjectID](mod.AuthorizeObjectsReadByID))
	return
}
