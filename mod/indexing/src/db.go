package indexing

import (
	"errors"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/indexing"
	"gorm.io/gorm"
)

type DB struct {
	*gorm.DB
}

func (db *DB) autoMigrate() error {
	return db.AutoMigrate(&dbRepoEntry{})
}

func newDB(gormDB *gorm.DB) (*DB, error) {
	db := &DB{DB: gormDB}

	err := db.DB.AutoMigrate()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (db *DB) addToRepo(repoName string, objectID *astral.ObjectID) (err error) {
	err = db.Transaction(func(tx *gorm.DB) error {
		// get next version
		var ver int
		err = tx.Model(&dbRepoEntry{}).
			Select("MAX(version)+1").
			Where("repo = ?", repoName, objectID).
			First(&ver).
			Error

		var row dbRepoEntry
		err = tx.Where("repo = ? and object_id = ?", repoName, objectID).First(&row).Error
		if err == nil {
			if row.Exist {
				return indexing.ErrObjectAlreadyAdded
			}
			row.Version = ver
			row.Exist = true
			return tx.Save(&row).Error
		}

		return tx.Create(&dbRepoEntry{
			Repo:     repoName,
			ObjectID: objectID,
			Version:  ver,
			Exist:    true,
		}).Error
	})

	return
}

func (db *DB) removeFromRepo(repoName string, objectID *astral.ObjectID) error {
	return db.Transaction(func(tx *gorm.DB) (err error) {
		var row dbRepoEntry
		err = tx.Where("repo = ? and object_id = ?", repoName, objectID).First(&row).Error
		switch {
		case err != nil:
			return err
		case !row.Exist:
			return errors.New("row already marked as removed")
		}

		err = tx.Model(&dbRepoEntry{}).
			Select("MAX(version)+1").
			Where("repo = ?", repoName, objectID).
			First(&row.Version).
			Error
		if err != nil {
			return err
		}

		row.Exist = false

		err = tx.Save(&row).Error

		return err
	})
}

func (db *DB) findExcessObjectIDs(repoName string, set []*astral.ObjectID) (excess []*astral.ObjectID, err error) {
	err = db.DB.
		Model(&dbRepoEntry{}).
		Select("object_id").
		Where("repo = ? and object_id not in (?) and exist = true", repoName, set).
		Find(&excess).
		Error

	return
}

func (db *DB) findMissingObjectIDs(repoName string, set []*astral.ObjectID) (missing []*astral.ObjectID, err error) {
	// fetch existing
	var existing []*astral.ObjectID
	err = db.DB.
		Model(&dbRepoEntry{}).
		Where("repo = ? and object_id in (?) and exist = true", repoName, set).
		Pluck("object_id", &existing).
		Error
	if err != nil {
		return nil, err
	}

	// prepare a lookup set
	present := make(map[string]struct{}, len(existing))
	for _, id := range existing {
		present[id.String()] = struct{}{}
	}

	// find missing
	for _, id := range set {
		if _, ok := present[id.String()]; !ok {
			missing = append(missing, id)
		}
	}

	return
}
