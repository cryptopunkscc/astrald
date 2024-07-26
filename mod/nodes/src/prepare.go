package nodes

import (
	"context"
)

func (mod *Module) Prepare(ctx context.Context) error {
	mod.Exonet.AddResolver(mod)

	return nil
}
