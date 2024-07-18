package apphost

import (
	"context"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/apphost"
)

func (mod *Module) Prepare(ctx context.Context) error {
	if i, err := mod.dir.Resolve(mod.config.DefaultIdentity); err != nil {
		mod.log.Errorv(1,
			"config: error resolving default identity %v: %v",
			mod.config.DefaultIdentity,
			err,
		)
	} else {
		mod.defaultID = i
	}

	// inject admin command
	if adm, err := core.Load[admin.Module](mod.node, admin.ModuleName); err == nil {
		adm.AddCommand(apphost.ModuleName, &Admin{mod: mod})
	}

	// load fixed access tokens from the config
	for token, name := range mod.config.Tokens {
		identity, err := mod.dir.Resolve(name)
		if err != nil {
			mod.log.Error("config: cannot resolve identity '%v': %v", name, err)
			continue
		}

		mod.db.Model(&dbAccessToken{}).Delete("token = ?", token)
		mod.db.Create(&dbAccessToken{
			Identity: identity.PublicKeyHex(),
			Token:    token,
		})
	}

	return nil
}
