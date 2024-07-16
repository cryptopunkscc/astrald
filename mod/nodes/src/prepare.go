package nodes

import (
	"context"
)

func (mod *Module) Prepare(ctx context.Context) error {
	mod.exonet.AddResolver(mod)

	return nil
}
