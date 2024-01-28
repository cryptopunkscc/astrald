package zip

import (
	"github.com/cryptopunkscc/astrald/data"
)

type dbZipContent struct {
	ZipID  string `gorm:"primaryKey"`
	Path   string `gorm:"primaryKey"`
	FileID string `gorm:"index"`
}

func (dbZipContent) TableName() string { return "zip_contents" }

func (mod *Module) dbFindByFileID(dataID data.ID) ([]dbZipContent, error) {
	var rows []dbZipContent

	tx := mod.db.Where("file_id = ?", dataID.String()).Find(&rows)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return rows, nil
}

func (mod *Module) dbFindByZipID(dataID data.ID) ([]dbZipContent, error) {
	var rows []dbZipContent

	tx := mod.db.Where("zip_id = ?", dataID.String()).Find(&rows)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return rows, nil
}
