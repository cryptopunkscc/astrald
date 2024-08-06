package apphost

import (
	"context"
)

func (mod *Module) Prepare(ctx context.Context) error {
	if i, err := mod.Dir.Resolve(mod.config.DefaultIdentity); err != nil {
		mod.log.Errorv(1,
			"config: error resolving default identity %v: %v",
			mod.config.DefaultIdentity,
			err,
		)
	} else {
		mod.defaultID = i
	}

	// load fixed access tokens from the config
	for token, name := range mod.config.Tokens {
		identity, err := mod.Dir.Resolve(name)
		if err != nil {
			mod.log.Error("config: cannot resolve identity '%v': %v", name, err)
			continue
		}

		mod.db.Model(&dbAccessToken{}).Delete("token = ?", token)
		mod.db.Create(&dbAccessToken{
			Identity: identity.String(),
			Token:    token,
		})
	}

	return nil
}
