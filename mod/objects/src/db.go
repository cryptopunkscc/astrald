package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"gorm.io/gorm"
)

type DB struct {
	*gorm.DB
}

func (db *DB) Migrate() error {
	return db.AutoMigrate(&dbObject{})
}

func (db *DB) Contains(id *astral.ObjectID) (b bool, err error) {
	err = db.
		Model(&dbObject{}).
		Where("id = ?", id).
		Select("count(*)>0").
		First(&b).Error
	return
}

func (db *DB) Find(id *astral.ObjectID) (row *dbObject, err error) {
	err = db.
		Where("id = ?", id).
		First(&row).Error
	return
}

func (db *DB) Create(id *astral.ObjectID, objectType string) (err error) {
	err = db.DB.Create(&dbObject{
		ID:   id,
		Type: objectType,
	}).Error
	return
}

func (db *DB) FindByType(objectType string) (rows []*dbObject, err error) {
	err = db.
		Where("type = ?", objectType).
		Find(&rows).Error
	return
}
