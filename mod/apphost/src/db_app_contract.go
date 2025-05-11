package apphost

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"github.com/cryptopunkscc/astrald/object"
	"time"
)

type dbAppContract struct {
	ObjectID  *object.ID       `gorm:"primaryKey"`
	AppID     *astral.Identity `gorm:"index"`
	HostID    *astral.Identity `gorm:"index"`
	StartsAt  time.Time        `gorm:"index"`
	ExpiresAt time.Time        `gorm:"index"`
}

func (dbAppContract) TableName() string { return apphost.DBPrefix + "app_contracts" }
