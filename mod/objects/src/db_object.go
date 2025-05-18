package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"time"
)

type dbObject struct {
	ID        *astral.ObjectID `gorm:"primaryKey"`
	Type      string           `gorm:"index"`
	CreatedAt time.Time        `gorm:"index"`
}

func (dbObject) TableName() string { return objects.DBPrefix + "objects" }
