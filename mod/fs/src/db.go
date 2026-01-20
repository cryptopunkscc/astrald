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
		Where("(updated_at != 0 OR mod_time != 0)").
		Select("count(*)>0").
		First(&b).Error
	return
}

func (db *DB) FindObject(pathPrefix string, objectID *astral.ObjectID) (rows []*dbLocalFile, err error) {
	err = db.
		Model(&dbLocalFile{}).
		Where("data_id = ?", objectID).
		Where("path like ?", pathPrefix+"%").
		Where("(updated_at != 0 OR mod_time != 0)").
		Find(&rows).
		Error

	return
}

func (db *DB) UniqueObjectIDs(pathPrefix string) (ids []*astral.ObjectID, err error) {
	err = db.
		Model(&dbLocalFile{}).
		Distinct("data_id").
		Where("path like ?", pathPrefix+"%").
		Where("(updated_at != 0 OR mod_time != 0)").
		Find(&ids).
		Error

	return
}

func (db *DB) FindByPath(path string) (row *dbLocalFile, err error) {
	err = db.
		Where("path = ?", path).
		Where("(updated_at != 0 OR mod_time != 0)").
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
		Where("(updated_at != 0 OR mod_time != 0)").
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
			lastID = row.ID
		}
	}
}

// InsertPaths batch inserts invalidated path records
func (db *DB) InsertPaths(paths []string) error {
	if len(paths) == 0 {
		return nil
	}

	records := make([]dbLocalFile, len(paths))
	for i, path := range paths {
		records[i] = dbLocalFile{
			Path:      path,
			ModTime:   time.Time{},
			UpdatedAt: time.Time{},
		}
	}

	return db.
		Clauses(clause.Insert{Modifier: "OR IGNORE"}).
		Create(&records).
		Error
}

// UpsertPath updates or inserts a path record (updates data_id/mod_time/updated_at)
func (db *DB) UpsertPath(
	path string,
	objectID *astral.ObjectID,
	modTime time.Time,
) error {
	now := time.Now()
	updated := &dbLocalFile{
		Path:      path,
		DataID:    objectID,
		ModTime:   modTime,
		UpdatedAt: now,
	}

	return db.
		Clauses(clause.OnConflict{
			DoUpdates: clause.AssignmentColumns([]string{
				"data_id",
				"mod_time",
				"updated_at",
			}),
		}).
		Create(updated).
		Error
}

// ValidatePath marks a path as valid (updates updated_at only)
func (db *DB) ValidatePath(path string) error {
	return db.Model(&dbLocalFile{}).
		Where("path = ?", path).
		Update("updated_at", time.Now()).Error
}

// InvalidateAllPaths marks all all paths for re-check
func (db *DB) InvalidateAllPaths() error {
	return db.Model(&dbLocalFile{}).
		Where("id > 0").
		UpdateColumn("updated_at", 0).
		Error
}

// Invalidate marks a path for re-check
func (db *DB) Invalidate(path string) (err error) {
	return db.Model(&dbLocalFile{}).
		Where("path = ?", path).
		Update("updated_at", 0).Error
}

// EachInvalidPath calls fn for each invalid path
func (db *DB) EachInvalidPath(fn func(string) error) error {
	const batchSize = 1000
	var lastID int64

	for {
		var rows []dbLocalFile

		err := db.
			Select("id, path").
			Where("(updated_at = 0 OR mod_time = 0)").
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
			lastID = row.ID
		}
	}
}
