package user

import (
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/mod/user"
	"time"
)

type dbContact struct {
	UserID    id.Identity `gorm:"primaryKey"`
	CreatedAt time.Time   `gorm:"index"`
}

func (dbContact) TableName() string { return user.DBPrefix + "contacts" }
