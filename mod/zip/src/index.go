package zip

import (
	_zip "archive/zip"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/zip"
)

const archiveSetPrefix = "zip:"

func (mod *Module) Index(zipID data.ID) error {
	var zipRow dbZip

	err := mod.db.Unscoped().Where("data_id = ?", zipID).First(&zipRow).Error
	if err == nil {
		if zipRow.DeletedAt.Time.IsZero() {
			return errors.New("already indexed")
		}
		return mod.restore(zipID)
	}

	// create a zip reader
	reader, err := _zip.NewReader(
		&readerAt{
			storage: mod.storage,
			dataID:  zipID,
		},
		int64(zipID.Size),
	)
	if err != nil {
		return fmt.Errorf("error reading zip file: %w", err)
	}

	mod.log.Logv(1, "indexing %v", zipID)

	// create database row and a set
	var setName = archiveSetPrefix + zipID.String()

	zipRow = dbZip{DataID: zipID, SetName: setName}
	err = mod.db.Create(&zipRow).Error
	if err != nil {
		return err
	}

	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue
		}

		f, err := file.Open()
		if err != nil {
			mod.log.Errorv(1, "open %v: %v", file.Name, err)
			continue
		}
		defer f.Close()

		fileID, err := data.ResolveAll(f)
		if err != nil {
			mod.log.Errorv(1, "resolve %v: %v", file.Name, err)
			continue
		}
		if fileID.Size == 0 {
			continue
		}

		err = mod.db.Model(&zipRow).
			Association("Contents").
			Append(&dbContents{
				Path:     file.Name,
				FileID:   fileID,
				Comment:  file.Comment,
				Modified: file.Modified,
			})
		if err != nil {
			return err
		}

		mod.log.Infov(1, "indexed %s (%v)", file.Name, fileID)
	}

	mod.events.Emit(zip.EventArchiveIndexed{DataID: zipID})

	return nil
}

func (mod *Module) Unindex(zipID data.ID) error {
	var row dbZip
	var err = mod.db.
		Model(&dbZip{}).
		Where("data_id = ?", zipID).
		First(&row).Error
	if err != nil {
		return err
	}

	return mod.db.Delete(&row).Error
}

func (mod *Module) restore(zipID data.ID) error {
	var tx = mod.db.
		Unscoped().
		Model(&dbZip{}).
		Where("data_id = ? and deleted_at is not null", zipID).
		Update("deleted_at", nil)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return errors.New("zip not deleted")
	}

	var setName string
	err := mod.db.
		Model(&dbZip{}).
		Select("set_name").
		Where("data_id = ?", zipID).
		First(&setName).Error

	return err
}

func (mod *Module) Forget(zipID data.ID) error {
	// find the row to be deleted
	var row dbZip
	var err = mod.db.
		Unscoped().
		Where("data_id = ?", zipID).
		First(&row).Error
	if err != nil {
		return fmt.Errorf("error fetching row: %w", err)
	}

	// delete contents first
	err = mod.db.
		Where("zip_id = ?", row.ID).
		Delete(&dbContents{}).Error
	if err != nil {
		return fmt.Errorf("error deleting contents: %w", err)
	}

	// delete the zip row
	err = mod.db.
		Unscoped().
		Delete(&row).Error
	if err != nil {
		return fmt.Errorf("error deleting row: %w", err)
	}

	return nil
}

func (mod *Module) isIndexed(zipID data.ID, includeDeleted bool) bool {
	var count int64
	db := mod.db
	if includeDeleted {
		db = db.Unscoped()
	}
	err := db.
		Model(&dbZip{}).
		Where("data_id = ?", zipID).
		Count(&count).Error
	if err != nil {
		mod.log.Errorv(2, "database error: %v", err)
		return false
	}

	return count > 0
}
