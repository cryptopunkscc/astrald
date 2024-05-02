package shares

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/shares"
	"github.com/cryptopunkscc/astrald/object"
)

type SelfAuthorizer struct {
	*Module
}

func (auth *SelfAuthorizer) Authorize(identity id.Identity, _ object.ID) error {
	if identity.IsEqual(auth.node.Identity()) {
		return nil
	}
	return shares.ErrDenied
}
