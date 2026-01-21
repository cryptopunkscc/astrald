package indexing

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/tree"
)

func (mod *Module) LoadDependencies() (err error) {
	ctx := astral.NewContext(nil)

	err = core.Inject(mod.node, &mod.Deps)
	if err != nil {
		return
	}

	mod.repos, err = tree.Query(ctx, mod.Tree.Root(), "/mod/indexing/repos", true)
	if err != nil {
		return err
	}

	mod.indexes, err = tree.Query(ctx, mod.Tree.Root(), "/mod/indexing/indexes", true)
	if err != nil {
		return err
	}

	return err
}
