package shares

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/shares"
)

type SelfAuthorizer struct {
	*Module
}

func (auth *SelfAuthorizer) Authorize(identity id.Identity, dataID data.ID) error {
	if identity.IsEqual(auth.node.Identity()) {
		return nil
	}
	return shares.ErrDenied
}
