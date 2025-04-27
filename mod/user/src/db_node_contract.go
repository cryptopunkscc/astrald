package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/user"
	"github.com/cryptopunkscc/astrald/object"
	"time"
)

type dbNodeContract struct {
	ObjectID  *object.ID       `gorm:"primaryKey"`
	UserID    *astral.Identity `gorm:"index"`
	NodeID    *astral.Identity `gorm:"index"`
	ExpiresAt time.Time        `gorm:"index"`
}

func (dbNodeContract) TableName() string { return user.DBPrefix + "node_contracts" }
