package admin

import (
	"context"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/mod/relay"
	"github.com/cryptopunkscc/astrald/node/modules"
)

func (mod *Module) Prepare(ctx context.Context) error {
	mod.relay, _ = modules.Load[relay.Module](mod.node, relay.ModuleName)
	mod.keys, _ = modules.Load[keys.Module](mod.node, keys.ModuleName)

	// load admins from config
	for _, name := range mod.config.Admins {
		adminID, err := mod.node.Resolver().Resolve(name)
		if err != nil {
			continue
		}

		mod.admins.Add(adminID.PublicKeyHex())
	}

	return nil
}
