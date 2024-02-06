package media

import (
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/media"
	"time"
)

type dbMediaInfo struct {
	DataID   data.ID `gorm:"primaryKey"`
	Type     string  `gorm:"index"`
	Artist   string  `gorm:"index"`
	Title    string  `gorm:"index"`
	Album    string  `gorm:"index"`
	Genre    string  `gorm:"index"`
	Duration time.Duration
}

func (dbMediaInfo) TableName() string {
	return media.DBPrefix + "media_info"
}
