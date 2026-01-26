package tree

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/ops"
	treecli "github.com/cryptopunkscc/astrald/mod/tree/client"
)

func (mod *Module) OpDelete(ctx *astral.Context, q *ops.Query, args treecli.DeleteArgs) (err error) {
	return treecli.NewNodeOps(mod.Root()).Delete(ctx, q, args)
}
