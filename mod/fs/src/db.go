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

func (db *DB) InTx(fn func(tx *DB) error) error {
	return db.Transaction(func(tx *gorm.DB) error {
		return fn(&DB{DB: tx})
	})
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

// EachPath calls fn for each path, using primary key pagination.
// If prefix is non-empty, only paths strictly under the prefix are matched (prefix+"/%"),
// not the prefix itself. This is correct for directory roots since only regular files
// are indexed, not directories.
func (db *DB) EachPath(prefix string, fn func(string) error) error {
	const batchSize = 1000
	var lastID int64

	for {
		var rows []dbLocalFile

		query := db.Select("id, path").Where("id > ?", lastID).Order("id ASC").Limit(batchSize)
		if prefix != "" {
			query = query.Where("path LIKE ?", prefix+"/%")
		}

		if err := query.Find(&rows).Error; err != nil {
			return err
		}

		if len(rows) == 0 {
			return nil
		}

		for _, row := range rows {
			if err := fn(row.Path); err != nil {
				return err
			}
			lastID = row.Id
		}
	}
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

func (db *DB) InvalidateAllPaths() error {
	return db.Model(&dbLocalFile{}).
		Where("id > 0").
		UpdateColumn("mod_time", 0).
		Error
}

func (db *DB) EachInvalidPath(fn func(string) error) error {
	const batchSize = 1000
	var lastID int64

	for {
		var rows []dbLocalFile

		err := db.
			Select("id, path").
			Where("mod_time = 0").
			Where("id > ?", lastID).
			Order("id ASC").
			Limit(batchSize).
			Find(&rows).
			Error
		if err != nil {
			return err
		}

		if len(rows) == 0 {
			return nil
		}

		for _, row := range rows {
			if err := fn(row.Path); err != nil {
				return err
			}
			lastID = row.Id
		}
	}
}
