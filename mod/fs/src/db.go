package fs

import (
	"github.com/cryptopunkscc/astrald/object"
	"gorm.io/gorm"
)

type DB struct {
	*gorm.DB
}

func (db *DB) ObjectExists(pathPrefix string, objectID *object.ID) (b bool, err error) {
	err = db.
		Model(&dbLocalFile{}).
		Where("data_id = ?", objectID).
		Where("path like ?", pathPrefix+"%").
		Select("count(*)>0").
		First(&b).Error
	return
}

func (db *DB) FindObject(pathPrefix string, objectID *object.ID) (rows []*dbLocalFile, err error) {
	err = db.
		Model(&dbLocalFile{}).
		Where("data_id = ?", objectID).
		Where("path like ?", pathPrefix+"%").
		Find(&rows).
		Error

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

func (db *DB) DeleteByPath(path string) (err error) {
	return db.
		Where("path = ?", path).
		Delete(&dbLocalFile{}).
		Error
}

func (db *DB) FindByObjectID(objectID *object.ID) (rows []*dbLocalFile, err error) {
	err = db.
		Model(&dbLocalFile{}).
		Where("data_id = ?", objectID).
		Find(&rows).
		Error

	return
}

func (db *DB) EachPath(fn func(string) error) (err error) {
	var batch = make([]*dbLocalFile, 1000)
	err = db.
		FindInBatches(&batch, 1000, func(tx *gorm.DB, n int) error {
			for _, row := range batch {
				err := fn(row.Path)
				if err != nil {
					return err
				}
			}
			return nil
		}).Error

	return
}
