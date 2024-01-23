package shares

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/shares"
)

type ACLAuthorizer struct {
	*Module
}

func (auth *ACLAuthorizer) Authorize(identity id.Identity, dataID data.ID) error {
	found, err := auth.localShareIndexContains(identity, dataID)
	if err != nil {
		return err
	}

	if !found {
		return shares.ErrDenied
	}

	return nil
}
