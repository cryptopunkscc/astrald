package acl

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"time"
)

type dbPerm struct {
	Identity  string    `gorm:"primaryKey,index"`
	DataID    string    `gorm:"primaryKey,index"`
	ExpiresAt time.Time `gorm:"index"`
	CreatedAt time.Time
}

func (dbPerm) TableName() string { return "perms" }

func (mod *Module) findPerm(identity id.Identity, dataID data.ID) *dbPerm {
	var dba dbPerm
	tx := mod.db.First(&dba,
		"identity = ? and data_id = ? and expires_at > ?",
		identity.String(),
		dataID.String(),
		time.Now(),
	)
	if tx.Error != nil {
		return nil
	}

	return &dbPerm{}
}
