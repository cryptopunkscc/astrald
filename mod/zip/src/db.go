package zip

import (
	_data "github.com/cryptopunkscc/astrald/data"
	"time"
)

type dbZipContent struct {
	ZipID     string
	Path      string
	FileID    string
	IndexedAt time.Time
}

func (dbZipContent) TableName() string { return "zip_contents" }

func (mod *Module) dbFindByFileID(dataID _data.ID) ([]dbZipContent, error) {
	var rows []dbZipContent

	tx := mod.db.Where("file_id = ?", dataID.String()).Find(&rows)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return rows, nil
}

func (mod *Module) dbFindByZipID(dataID _data.ID) ([]dbZipContent, error) {
	var rows []dbZipContent

	tx := mod.db.Where("zip_id = ?", dataID.String()).Find(&rows)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return rows, nil
}
