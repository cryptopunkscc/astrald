package services

import (
	"github.com/cryptopunkscc/astrald/astral"
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

// RemoveService marks a service as disabled
func (db *DB) RemoveService(name astral.String8, identity *astral.Identity) error {
	return db.
		Model(&dbService{}).
		Where("name = ? AND identity = ?", name, identity).
		Update("enabled", false).
		Error
}

// ServiceExists checks if a service exists and is enabled
func (db *DB) ServiceExists(name astral.String8, identity *astral.Identity) (exists bool) {
	db.
		Model(&dbService{}).
		Select("1").
		Where("name = ? AND identity = ? AND enabled = ?", name, identity, true).
		Limit(1).
		Scan(&exists)
	return
}

// ClearDisabledServices removes old disabled service entries (cleanup)
func (db *DB) ClearDisabledServices() error {
	return db.
		Where("enabled = ?", false).
		Delete(&dbService{}).
		Error
}
