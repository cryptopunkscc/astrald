package fs

import (
	"github.com/cryptopunkscc/astrald/mod/fs"
	"github.com/cryptopunkscc/astrald/object"
	"time"
)

type dbLocalFile struct {
	Path      string     `gorm:"primaryKey"`
	DataID    *object.ID `gorm:"index"`
	ModTime   time.Time
	UpdatedAt time.Time
}

func (dbLocalFile) TableName() string { return fs.DBPrefix + "local_files" }
