package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/user"
	"time"
)

type dbContact struct {
	UserID    *astral.Identity `gorm:"primaryKey"`
	CreatedAt time.Time        `gorm:"index"`
}

func (dbContact) TableName() string { return user.DBPrefix + "contacts" }
