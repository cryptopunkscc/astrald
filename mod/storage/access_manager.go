package storage

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	storage "github.com/cryptopunkscc/astrald/mod/storage/api"
	"sync"
	"time"
)

var _ storage.AccessManager = &AccessManager{}

type AccessManager struct {
	*Module
	verifiers map[storage.AccessVerifier]struct{}
	mu        sync.Mutex
}

type Access struct {
	Identity  id.Identity
	DataID    data.ID
	ExpiresAt time.Time
}

type dbAccess struct {
	Identity  string    `gorm:"primaryKey,index"`
	DataID    string    `gorm:"primaryKey,index"`
	ExpiresAt time.Time `gorm:"index"`
	CreatedAt time.Time
}

func (dbAccess) TableName() string { return "accesses" }

func NewAccessManager(module *Module) *AccessManager {
	return &AccessManager{
		Module:    module,
		verifiers: make(map[storage.AccessVerifier]struct{}, 0),
	}
}

func (mod *AccessManager) Grant(identity id.Identity, dataID data.ID, expiresAt time.Time) error {
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

func (mod *AccessManager) Revoke(identity id.Identity, dataID data.ID) error {
	return mod.db.Delete(&dbAccess{}, "identity = ? and data_id = ?", identity.String(), dataID.String()).Error
}

func (mod *AccessManager) Verify(identity id.Identity, dataID data.ID) bool {
	// local node has access to everything
	if identity.IsEqual(mod.node.Identity()) {
		return true
	}

	// check if the data is public
	if a := mod.findAccess(id.Anyone, dataID); a != nil {
		return true
	}

	// check if the identity has access
	if !identity.IsZero() {
		if a := mod.findAccess(identity, dataID); a != nil {
			return true
		}
	}

	for ac := range mod.verifiers {
		if ac.Verify(identity, dataID) {
			return true
		}
	}

	return false
}

func (mod *AccessManager) AddAccessVerifier(checker storage.AccessVerifier) {
	mod.mu.Lock()
	defer mod.mu.Unlock()

	mod.verifiers[checker] = struct{}{}
}

func (mod *AccessManager) RemoveAccessVerifier(checker storage.AccessVerifier) {
	mod.mu.Lock()
	defer mod.mu.Unlock()

	delete(mod.verifiers, checker)
}

func (mod *AccessManager) findAccess(identity id.Identity, dataID data.ID) *Access {
	var dba dbAccess
	tx := mod.db.First(&dba,
		"identity = ? and data_id = ? and expires_at > ?",
		identity.String(),
		dataID.String(),
		time.Now(),
	)
	if tx.Error != nil {
		return nil
	}

	return &Access{
		Identity:  identity,
		DataID:    dataID,
		ExpiresAt: dba.ExpiresAt,
	}
}

func (dba dbAccess) toAccess() (*Access, error) {
	var err error
	var a = &Access{
		ExpiresAt: dba.ExpiresAt,
	}
	if dba.Identity != "" {
		if a.Identity, err = id.ParsePublicKeyHex(dba.Identity); err != nil {
			return nil, err
		}
	}
	if a.DataID, err = data.Parse(dba.DataID); err != nil {
		return nil, err
	}

	return a, nil
}

func toDbAccess(access *Access) dbAccess {
	return dbAccess{
		Identity:  access.Identity.String(),
		DataID:    access.DataID.String(),
		ExpiresAt: access.ExpiresAt,
	}
}
