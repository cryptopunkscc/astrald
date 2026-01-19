package fs

import (
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/fs"
)

type dbLocalFile struct {
	Id        int64            `gorm:"primaryKey;autoIncrement"`
	Path      string           `gorm:"index"`
	DataID    *astral.ObjectID `gorm:"index"`
	ModTime   time.Time
	UpdatedAt time.Time
}

func (dbLocalFile) TableName() string { return fs.DBPrefix + "local_files" }
