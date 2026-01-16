package fs

import (
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DB struct {
	*gorm.DB
}

func (db *DB) ObjectExists(pathPrefix string, objectID *astral.ObjectID) (b bool, err error) {
	err = db.
		Model(&dbLocalFile{}).
		Where("data_id = ?", objectID).
		Where("path like ?", pathPrefix+"%").
		Select("count(*)>0").
		First(&b).Error
	return
}

func (db *DB) FindObject(pathPrefix string, objectID *astral.ObjectID) (rows []*dbLocalFile, err error) {
	err = db.
		Model(&dbLocalFile{}).
		Where("data_id = ?", objectID).
		Where("path like ?", pathPrefix+"%").
		Find(&rows).
		Error

	return
}

func (db *DB) UniqueObjectIDs(pathPrefix string) (ids []*astral.ObjectID, err error) {
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

func (db *DB) FindByObjectID(objectID *astral.ObjectID) (rows []*dbLocalFile, err error) {
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

func (db *DB) UpsertPath(
	path string,
	objectID *astral.ObjectID,
	modTime time.Time,
) error {
	updated := &dbLocalFile{
		Path:    path,
		DataID:  objectID,
		ModTime: modTime,
	}

	return db.
		Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns([]string{
				"data_id",
				"mod_time",
			}),
		}).
		Create(updated).
		Error
}

func (db *DB) InvalidatePath(path string) (err error) {
	return db.Model(&dbLocalFile{}).
		Where("path = ?", path).
		Update("mod_time", 0).Error
}

func (db *DB) InvalidateAll() (err error) {
	return db.Model(&dbLocalFile{}).Update("mod_time", 0).Error
}
