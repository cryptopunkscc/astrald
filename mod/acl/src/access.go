package acl

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
)

func (mod *Module) Grant(identity id.Identity, dataID data.ID) error {
	return mod.addToLocalShareIndex(identity, dataID)
}

func (mod *Module) Revoke(identity id.Identity, dataID data.ID) error {
	return mod.removeFromLocalShareIndex(identity, dataID)
}

func (mod *Module) Verify(identity id.Identity, dataID data.ID) bool {
	found, err := mod.localShareIndexContains(identity, dataID)

	return (err == nil) && found
}
