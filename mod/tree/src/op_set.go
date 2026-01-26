package tree

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/ops"
	treecli "github.com/cryptopunkscc/astrald/mod/tree/client"
)

func (mod *Module) OpSet(ctx *astral.Context, q *ops.Query, args treecli.SetArgs) (err error) {
	return treecli.NewNodeOps(mod.Root()).Set(ctx, q, args)
}
