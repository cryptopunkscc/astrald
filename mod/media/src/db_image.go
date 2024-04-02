package media

import (
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/media"
)

type dbImage struct {
	DataID data.ID `gorm:"primaryKey"`
	Format string  `gorm:"index"`
	Width  int     `gorm:"index"`
	Height int     `gorm:"index"`
}

func (dbImage) TableName() string { return media.DBPrefix + "images" }
