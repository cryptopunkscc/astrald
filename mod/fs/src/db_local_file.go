package fs

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/fs"
	"time"
)

type dbLocalFile struct {
	Path      string           `gorm:"primaryKey"`
	DataID    *astral.ObjectID `gorm:"index"`
	ModTime   time.Time
	UpdatedAt time.Time
}

func (dbLocalFile) TableName() string { return fs.DBPrefix + "local_files" }
