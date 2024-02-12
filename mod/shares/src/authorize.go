package shares

import (
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/shares"
)

func (mod *Module) Authorize(identity id.Identity, dataID data.ID) error {
	for _, authorizer := range mod.authorizers.Clone() {
		var err = authorizer.Authorize(identity, dataID)
		switch {
		case err == nil:
			return nil
		case errors.Is(err, shares.ErrDenied):
		default:
			return err
		}
	}
	return shares.ErrDenied
}

func (mod *Module) addAuthorizer(authorizer DataAuthorizer) error {
	return mod.authorizers.Add(authorizer)
}

func (mod *Module) removeAuthorizer(authorizer DataAuthorizer) error {
	return mod.authorizers.Remove(authorizer)
}
