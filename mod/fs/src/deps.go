package fs

import (
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type Deps struct {
	Auth    auth.Module
	Dir     dir.Module
	Objects objects.Module
	Shell   shell.Module
}

func (mod *Module) LoadDependencies() (err error) {
	err = core.Inject(mod.node, &mod.Deps)
	if err != nil {
		return
	}

	// add preconfigured repos
	for _, repo := range mod.repos.Clone() {
		mod.Objects.AddRepository(repo.Label(), repo)
	}

	return
}
