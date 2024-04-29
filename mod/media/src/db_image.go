package media

import (
	"github.com/cryptopunkscc/astrald/mod/media"
	"github.com/cryptopunkscc/astrald/object"
)

type dbImage struct {
	DataID object.ID `gorm:"primaryKey"`
	Format string    `gorm:"index"`
	Width  int       `gorm:"index"`
	Height int       `gorm:"index"`
}

func (dbImage) TableName() string { return media.DBPrefix + "images" }
