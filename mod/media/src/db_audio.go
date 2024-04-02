package media

import (
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/media"
	"time"
)

type dbAudio struct {
	DataID   data.ID `gorm:"primaryKey"`
	Format   string  `gorm:"index"`
	Duration time.Duration
	Title    string `gorm:"index"`
	Artist   string `gorm:"index"`
	Album    string `gorm:"index"`
	Genre    string `gorm:"index"`
	Year     int    `gorm:"index"`
}

func (dbAudio) TableName() string { return media.DBPrefix + "audio" }
