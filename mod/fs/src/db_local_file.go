package fs

import (
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/fs"
	"time"
)

type dbLocalFile struct {
	Path      string  `gorm:"primaryKey,index"`
	DataID    data.ID `gorm:"index"`
	IndexedAt time.Time
}

func (dbLocalFile) TableName() string { return fs.DBPrefix + "local_files" }

func (mod *Module) dbFindByPath(path string) *dbLocalFile {
	var row dbLocalFile

	tx := mod.db.First(&row, "path = ?", path)
	if tx.Error != nil {
		return nil
	}

	return &row
}

func (mod *Module) dbFindByID(dataID data.ID) []*dbLocalFile {
	var rows []*dbLocalFile

	tx := mod.db.Where("data_id = ?", dataID.String()).Find(&rows)
	if tx.Error != nil {
		return nil
	}

	return rows
}

func (mod *Module) dbDeleteByPath(path string) error {
	return mod.db.Where("path = ?", path).Delete(&dbLocalFile{}).Error
}
