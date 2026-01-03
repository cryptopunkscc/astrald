package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/astrald"
)

// Push pushes the object to the target node. The sender node is derived from the context.
func (mod *Module) Push(ctx *astral.Context, targetID *astral.Identity, obj astral.Object) (err error) {
	if targetID.IsEqual(mod.node.Identity()) {
		return mod.Receive(obj, ctx.Identity())
	}

	return astrald.NewObjectsClient(targetID, nil).Push(ctx, obj)
}
