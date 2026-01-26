package tree

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/ops"
	treecli "github.com/cryptopunkscc/astrald/mod/tree/client"
)

func (mod *Module) OpGet(ctx *astral.Context, q *ops.Query, args treecli.GetArgs) (err error) {
	return treecli.NewNodeOps(mod.Root()).Get(ctx, q, args)
}
