package user

import (
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/mod/user"
	"gorm.io/gorm"
)

type dbIdentity struct {
	Identity id.Identity `gorm:"primaryKey"`
}

func (dbIdentity) TableName() string {
	return user.DBPrefix + "identities"
}

func (mod *Module) loadUserID() (id.Identity, error) {
	var row dbIdentity

	err := mod.db.First(&row).Error

	return row.Identity, err
}

func (mod *Module) storeUserID(userID id.Identity) error {
	var err = mod.db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&dbIdentity{}).Error
	if err != nil {
		return err
	}

	return mod.db.
		Create(&dbIdentity{Identity: userID}).
		Error
}
