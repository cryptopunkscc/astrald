package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
)

// Push pushes the object to the target node. The sender node is derived from the context.
func (mod *Module) Push(ctx *astral.Context, nodeID *astral.Identity, obj astral.Object) (err error) {
	if nodeID.IsEqual(mod.node.Identity()) {
		return mod.Receive(obj, ctx.Identity())
	}

	c, err := mod.On(nodeID, ctx.Identity())
	if err != nil {
		return err
	}

	return c.Push(ctx, obj)
}
