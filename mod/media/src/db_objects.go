package media

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/media"
)

type dbObject struct {
	ObjectID *astral.ObjectID `gorm:"primaryKey"`
	Audio    *dbAudio         `gorm:"foreignKey:ObjectID"`
}

func (dbObject) TableName() string { return media.DBPrefix + "objects" }
