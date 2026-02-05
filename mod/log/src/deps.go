package log

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/log"
	"github.com/cryptopunkscc/astrald/mod/tree"
)

func (mod *Module) LoadDependencies(ctx *astral.Context) (err error) {
	err = core.Inject(mod.node, &mod.Deps)
	if err != nil {
		return
	}

	log.IdentityResolver.Set(mod.Dir)

	// bind the config
	err = tree.BindPath(ctx, &mod.config, mod.Tree.Root(), "/mod/log/config", true)
	if err != nil {
		return err
	}

	return
}
