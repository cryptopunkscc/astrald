package objects

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
)

func (mod *Module) Push(ctx context.Context, source *astral.Identity, target *astral.Identity, obj astral.Object) (err error) {
	if source.IsZero() {
		source = mod.node.Identity()
	}

	if target.IsEqual(mod.node.Identity()) {
		return mod.Receive(obj, source)
	}

	c, err := mod.On(target, source)
	if err != nil {
		return err
	}

	return c.Push(ctx, obj)
}
