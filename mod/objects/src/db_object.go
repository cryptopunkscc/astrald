package objects

import (
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
	"time"
)

type dbObject struct {
	ID        *object.ID `gorm:"primaryKey"`
	Type      string     `gorm:"index"`
	CreatedAt time.Time  `gorm:"index"`
}

func (dbObject) TableName() string { return objects.DBPrefix + "objects" }
