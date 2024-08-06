package admin

import (
	"context"
	"github.com/cryptopunkscc/astrald/core"
)

func (mod *Module) Prepare(ctx context.Context) (err error) {
	err = core.Inject(mod.node, &mod.Deps)
	if err != nil {
		return
	}

	// load admins from config
	for _, name := range mod.config.Admins {
		adminID, err := mod.Dir.Resolve(name)
		if err != nil {
			continue
		}

		mod.admins.Add(adminID.String())
	}

	mod.Auth.AddAuthorizer(mod)

	return nil
}
