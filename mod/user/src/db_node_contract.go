package user

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/user"
	"github.com/cryptopunkscc/astrald/object"
	"time"
)

type dbNodeContract struct {
	ObjectID  object.ID   `gorm:"primaryKey"`
	UserID    id.Identity `gorm:"index"`
	NodeID    id.Identity `gorm:"index"`
	ExpiresAt time.Time   `gorm:"index"`
}

func (dbNodeContract) TableName() string { return user.DBPrefix + "node_contracts" }
