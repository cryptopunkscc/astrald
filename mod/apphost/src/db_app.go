package apphost

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/apphost"
)

type dbApp struct {
	AppID    string           `gorm:"primaryKey"`
	Identity *astral.Identity `gorm:"index"`
}

func (dbApp) TableName() string { return apphost.DBPrefix + "apps" }
