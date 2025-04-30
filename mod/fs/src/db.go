package fs

import (
	"github.com/cryptopunkscc/astrald/object"
	"gorm.io/gorm"
)

type DB struct {
	*gorm.DB
}

func (db *DB) ObjectExists(objectID *object.ID) (b bool, err error) {
	err = db.
		Model(&dbLocalFile{}).
		Where("data_id = ?", objectID).
		Select("count(*)>0").
		First(&b).Error
	return
}

func (db *DB) UniqueObjectIDs(pathPrefix string) (ids []*object.ID, err error) {
	err = db.
		Model(&dbLocalFile{}).
		Distinct("data_id").
		Where("path like ?", pathPrefix+"%").
		Find(&ids).
		Error

	return
}

func (db *DB) FindByPath(path string) (row *dbLocalFile, err error) {
	err = db.
		Where("path = ?", path).
		First(&row).
		Error

	return
}

func (db *DB) FindByObjectID(objectID *object.ID) (rows []*dbLocalFile, err error) {
	err = db.
		Model(&dbLocalFile{}).
		Where("data_id = ?", objectID).
		Find(&rows).
		Error

	return
}
