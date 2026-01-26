package tree

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/ops"
	treecli "github.com/cryptopunkscc/astrald/mod/tree/client"
)

func (mod *Module) OpList(ctx *astral.Context, q *ops.Query, args treecli.ListArgs) (err error) {
	return treecli.NewNodeOps(mod.Root()).List(ctx, q, args)
}
