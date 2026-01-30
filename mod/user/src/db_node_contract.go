package user

import (
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/user"
)

type dbSignedNodeContract struct {
	ObjectID  *astral.ObjectID `gorm:"primaryKey"`
	UserID    *astral.Identity `gorm:"index"`
	NodeID    *astral.Identity `gorm:"index"`
	ExpiresAt time.Time        `gorm:"index"`
	StartsAt  time.Time        `gorm:"index"`
}

func (dbSignedNodeContract) TableName() string { return user.DBPrefix + "signed_node_contracts" }
