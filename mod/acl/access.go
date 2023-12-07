package acl

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"time"
)

func (mod *Module) Grant(identity id.Identity, dataID data.ID, expiresAt time.Time) error {
	var dba dbPerm

	// check if the user already has permission to access the data
	var tx = mod.db.First(&dba, "identity = ? and data_id = ?", identity.String(), dataID.String())

	if tx.Error == nil {
		if dba.ExpiresAt.Before(expiresAt) {
			return mod.db.Model(&dbPerm{}).
				Where("identity = ? and data_id = ?", identity.String(), dataID.String()).
				Update("expires_at", expiresAt).Error
		}
		return nil
	}

	return mod.db.Create(&dbPerm{
		Identity:  identity.String(),
		DataID:    dataID.String(),
		ExpiresAt: expiresAt,
	}).Error
}

func (mod *Module) Revoke(identity id.Identity, dataID data.ID) error {
	return mod.db.Delete(
		&dbPerm{},
		"identity = ? and data_id = ?",
		identity.String(),
		dataID.String(),
	).Error
}

func (mod *Module) Verify(identity id.Identity, dataID data.ID) bool {
	// check if the data is public
	if perm := mod.findPerm(id.Anyone, dataID); perm != nil {
		return true
	}

	// check if the identity has permission
	if !identity.IsZero() {
		if perm := mod.findPerm(identity, dataID); perm != nil {
			return true
		}
	}

	return false
}
