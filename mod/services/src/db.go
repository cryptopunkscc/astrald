package services

import (
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/services"
	"gorm.io/gorm"
)

type DB struct {
	*gorm.DB
}

func (db *DB) Migrate() error {
	return db.AutoMigrate(
		&dbService{},
	)
}

// FIXME: upsert

// InsertService creates a new service record and expires previous ones for the same name+identity
func (db *DB) InsertService(svc services.Service, expiresIn astral.Duration) error {
	now := astral.Now()
	expiresAt := astral.Time(now.Time().Add(time.Duration(expiresIn)))

	if err := db.
		Model(&dbService{}).
		Where("name = ? AND identity = ? AND expires_at > ?", svc.Name, svc.Identity, now).
		Update("expires_at", now).
		Error; err != nil {
		return err
	}

	return db.Create(&dbService{
		Name:        svc.Name,
		Identity:    svc.Identity,
		Composition: svc.Composition,
		ExpiresAt:   expiresAt,
	}).Error
}

// GetIdentityServices returns all current services for a specific identity
func (db *DB) GetIdentityServices(identity *astral.Identity) ([]dbService, error) {
	var svcList []dbService
	now := astral.Now()
	err := db.
		Where("identity = ? AND expires_at > ?", identity, now).
		Order("created_at DESC").
		Find(&svcList).
		Error
	return svcList, err
}

func (db *DB) InvalidateServices(identity *astral.Identity) error {
	now := astral.Now()

	return db.
		Model(&dbService{}).
		Where("identity = ? AND expires_at > ?", identity, now).
		Update("expires_at", now).
		Error
}

// InvalidateService expires a specific service for an identity
func (db *DB) InvalidateService(name astral.String8, identity *astral.Identity) error {
	now := astral.Now()

	return db.
		Model(&dbService{}).
		Where("name = ? AND identity = ? AND expires_at > ?", name, identity, now).
		Update("expires_at", now).
		Error
}
