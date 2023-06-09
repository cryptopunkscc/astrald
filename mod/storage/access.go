package storage

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"time"
)

type Access struct {
	Identity  id.Identity
	DataID    data.ID
	ExpiresAt time.Time
}

func (mod *Module) FindAccess(identity id.Identity, dataID data.ID) *Access {
	var dba dbAccess
	tx := mod.db.First(&dba, "identity = ? and data_id = ?", identity.String(), dataID.String())
	if tx.Error != nil {
		return nil
	}
	return &Access{
		Identity:  identity,
		DataID:    dataID,
		ExpiresAt: dba.ExpiresAt,
	}
}

func (mod *Module) RevokeAccess(identity id.Identity, dataID data.ID) error {
	return mod.db.Delete(&dbAccess{}, "identity = ? and data_id = ?", identity.String(), dataID.String()).Error
}

func (mod *Module) GrantAccess(identity id.Identity, dataID data.ID, expiresAt time.Time) error {
	var dba dbAccess
	tx := mod.db.First(&dba, "identity = ? and data_id = ?", identity.String(), dataID.String())
	if tx.Error == nil {
		if dba.ExpiresAt.Before(expiresAt) {
			return mod.db.Model(&dbAccess{}).
				Where("identity = ? and data_id = ?", identity.String(), dataID.String()).
				Update("expires_at", expiresAt).Error
		}
		return nil
	}

	return mod.db.Create(&dbAccess{
		Identity:  identity.String(),
		DataID:    dataID.String(),
		ExpiresAt: expiresAt,
	}).Error
}
