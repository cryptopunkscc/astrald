package media

import (
	"github.com/cryptopunkscc/astrald/mod/media"
	"github.com/cryptopunkscc/astrald/object"
	"time"
)

type dbVideo struct {
	DataID   object.ID     `gorm:"primaryKey"`
	Format   string        `gorm:"index"`
	Title    string        `gorm:"index"`
	Duration time.Duration `gorm:"index"`
}

func (dbVideo) TableName() string { return media.DBPrefix + "videos" }
