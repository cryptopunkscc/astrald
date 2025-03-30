package kos

import (
	"github.com/cryptopunkscc/astrald/astral"
	"gorm.io/gorm"
)

type DB struct {
	*gorm.DB
}

func (db *DB) Set(id *astral.Identity, key string, otype string, payload []byte) error {
	// First try to find existing record
	existing := dbEntry{}
	result := db.Where("identity = ? AND key = ?", id, key).First(&existing)

	switch result.Error {
	case nil:
		return db.Model(&existing).Updates(map[string]any{
			"type":    otype,
			"payload": payload,
		}).Error

	case gorm.ErrRecordNotFound:
		return db.Create(&dbEntry{
			Identity: id,
			Key:      key,
			Type:     otype,
			Payload:  payload,
		}).Error
	}

	return result.Error
}

func (db *DB) Get(id *astral.Identity, key string) (string, []byte, error) {
	var row dbEntry
	var err = db.Where("identity = ? AND key = ?", id, key).First(&row).Error

	return row.Type, row.Payload, err
}

func (db *DB) Delete(id *astral.Identity, key string) error {
	return db.Where("identity = ? AND key = ?", id, key).Delete(&dbEntry{}).Error
}
