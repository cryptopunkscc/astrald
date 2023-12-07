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

func (mod *Module) dbFindByID(dataID _data.ID) *dbZipContent {
	var row *dbZipContent

	tx := mod.db.Where("file_id = ?", dataID.String()).Find(&row)
	if tx.Error != nil {
		return nil
	}

	return row
}
