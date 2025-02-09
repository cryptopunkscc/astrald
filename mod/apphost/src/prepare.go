package apphost

import (
	"context"
	"time"
)

func (mod *Module) Prepare(ctx context.Context) error {
	// load fixed access tokens from the config
	for token, name := range mod.config.Tokens {
		identity, err := mod.Dir.ResolveIdentity(name)
		if err != nil {
			mod.log.Error("config: cannot resolve identity '%v': %v", name, err)
			continue
		}

		mod.db.Model(&dbAccessToken{}).Delete("token = ?", token)
		mod.db.Create(&dbAccessToken{
			Identity:  identity,
			Token:     token,
			ExpiresAt: time.Now().Add(time.Hour * 24 * 365 * 100),
		})
	}

	return nil
}
