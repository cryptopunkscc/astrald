package media

import (
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/media"
	"time"
)

type dbVideo struct {
	DataID   data.ID       `gorm:"primaryKey"`
	Format   string        `gorm:"index"`
	Title    string        `gorm:"index"`
	Duration time.Duration `gorm:"index"`
}

func (dbVideo) TableName() string { return media.DBPrefix + "videos" }
