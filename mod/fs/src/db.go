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
