package user

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/user"
)

type dbIdentity struct {
	Identity id.Identity `gorm:"primaryKey"`
	SetName  string      `gorm:"uniqueIndex"`
}

func (dbIdentity) TableName() string {
	return user.DBPrefix + "identities"
}
