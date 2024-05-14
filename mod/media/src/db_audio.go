package media

import (
	"github.com/cryptopunkscc/astrald/mod/media"
	"github.com/cryptopunkscc/astrald/object"
	"time"
)

type dbAudio struct {
	ObjectID object.ID `gorm:"primaryKey"`
	Format   string    `gorm:"index"`
	Duration time.Duration
	Title    string `gorm:"index"`
	Artist   string `gorm:"index"`
	Album    string `gorm:"index"`
	Genre    string `gorm:"index"`
	Year     int    `gorm:"index"`
}

func (dbAudio) TableName() string { return media.DBPrefix + "audio" }
