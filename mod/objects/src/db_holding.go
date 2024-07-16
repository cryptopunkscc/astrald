package objects

import (
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
	"time"
)

type dbHolding struct {
	HolderID  id.Identity `gorm:"primaryKey"`
	ObjectID  object.ID   `gorm:"primaryKey"`
	CreatedAt time.Time
}

func (dbHolding) TableName() string { return objects.DBPrefix + "holdings" }
