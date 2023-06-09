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

type AccessChecker interface {
	CheckAccess(identity id.Identity, dataID data.ID) bool
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

func (mod *Module) CheckAccess(identity id.Identity, dataID data.ID) bool {
	if identity.IsZero() {
		return false
	}
	if identity.IsEqual(mod.node.Identity()) {
		return true
	}
	if a := mod.FindAccess(identity, dataID); a != nil {
		if a.ExpiresAt.After(time.Now()) {
			return true
		}
	}

	for ac := range mod.accessCheckers {
		if ac.CheckAccess(identity, dataID) {
			return true
		}
	}

	return false
}

func (mod *Module) AddAccessChecker(checker AccessChecker) {
	mod.accessCheckersMu.Lock()
	defer mod.accessCheckersMu.Unlock()

	mod.accessCheckers[checker] = struct{}{}
}

func (mod *Module) RemoveAccessChecker(checker AccessChecker) {
	mod.accessCheckersMu.Lock()
	defer mod.accessCheckersMu.Unlock()

	delete(mod.accessCheckers, checker)
}

func (mod *Module) DataAccessCountByIdentity(identity id.Identity) int {
	if identity.IsZero() {
		return 0
	}

	var c int64
	mod.db.Model(&dbAccess{}).Where("identity = ?", identity.String()).Count(&c)
	return int(c)
}
