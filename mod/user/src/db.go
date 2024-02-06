package user

import "github.com/cryptopunkscc/astrald/mod/user"

type dbIdentity struct {
	Identity string `gorm:"primaryKey"`
}

func (dbIdentity) TableName() string {
	return user.DBPrefix + "identities"
}
