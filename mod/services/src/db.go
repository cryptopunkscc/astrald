package services

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/services"
	"gorm.io/gorm"
)

type DB struct {
	db *gorm.DB
}

func (db *DB) Migrate() error {
	return db.db.AutoMigrate(
		&dbService{},
	)
}

// InTx executes fn inside a database transaction.
// This is the only transaction API - commit/rollback are fully encapsulated.
func (db *DB) InTx(fn func(tx *DB) error) error {
	return db.db.Transaction(func(txGorm *gorm.DB) error {
		return fn(&DB{db: txGorm})
	})
}

// RemoveIdentityServices deletes all cached services for a specific identity
func (db *DB) RemoveIdentityServices(identity *astral.Identity) error {
	return db.db.
		Delete(&dbService{}, "identity = ?", identity).
		Error
}

// RemoveIdentityService deletes a specific cached service for a specific identity
func (db *DB) RemoveIdentityService(name astral.String8, identity *astral.Identity) error {
	return db.db.
		Delete(&dbService{}, "name = ? AND identity = ?", name, identity).
		Error
}

// InsertService creates a new service record
func (db *DB) InsertService(svc *services.Service) error {
	return db.db.Create(&dbService{
		Name:        svc.Name,
		Identity:    svc.Identity,
		Composition: svc.Composition,
	}).Error
}

// GetIdentityServices returns all current services for a specific identity
func (db *DB) GetIdentityServices(identity *astral.Identity) ([]dbService, error) {
	var svcList []dbService
	err := db.db.
		Where("identity = ?", identity).
		Order("created_at DESC").
		Find(&svcList).
		Error
	return svcList, err
}
