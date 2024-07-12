package presence

import (
	"context"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/admin"
)

func (mod *Module) Prepare(ctx context.Context) (err error) {
	// inject admin command
	if adm, err := core.Load[admin.Module](mod.node, admin.ModuleName); err == nil {
		adm.AddCommand(ModuleName, NewAdmin(mod))
	}

	return nil
}
