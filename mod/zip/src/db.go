package zip

import (
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/zip"
	"gorm.io/gorm"
	"time"
)

type dbZip struct {
	ID        uint         `gorm:"primarykey"`
	DataID    data.ID      `gorm:"uniqueIndex"`
	Contents  []dbContents `gorm:"OnDelete:CASCADE;foreignKey:ZipID"`
	CreatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (dbZip) TableName() string { return zip.DBPrefix + "zips" }

type dbContents struct {
	ZipID    uint `gorm:"primaryKey"`
	Zip      *dbZip
	FileID   data.ID `gorm:"index"`
	Path     string  `gorm:"primaryKey"`
	Comment  string
	Modified time.Time
}

func (dbContents) TableName() string { return zip.DBPrefix + "contents" }

func (mod *Module) dbFindByFileID(dataID data.ID) ([]dbContents, error) {
	var rows []dbContents

	tx := mod.db.
		Unscoped().
		Preload("Zip").
		Where("file_id = ?", dataID).
		Find(&rows)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return rows, nil
}

func (mod *Module) dbFindByZipID(zipID data.ID) ([]dbContents, error) {
	var rows []dbContents

	tx := mod.db.Where("zip_id = ?", zipID).Find(&rows)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return rows, nil
}
