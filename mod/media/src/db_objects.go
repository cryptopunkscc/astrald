package media

import (
	"github.com/cryptopunkscc/astrald/mod/media"
	"github.com/cryptopunkscc/astrald/object"
)

type dbObject struct {
	ObjectID *object.ID `gorm:"primaryKey"`
	Audio    *dbAudio   `gorm:"foreignKey:ObjectID"`
}

func (dbObject) TableName() string { return media.DBPrefix + "objects" }
