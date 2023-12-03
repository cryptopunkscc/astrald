package storage

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"time"
)

type dbAccess struct {
	Identity  string    `gorm:"primaryKey,index"`
	DataID    string    `gorm:"primaryKey,index"`
	ExpiresAt time.Time `gorm:"index"`
	CreatedAt time.Time
}

func (dbAccess) TableName() string { return "accesses" }

func (mod *Module) setupDatabase() (err error) {
	// Migrate the schema
	if err := mod.db.AutoMigrate(&dbAccess{}); err != nil {
		return err
	}

	if err := mod.db.AutoMigrate(&dbFile{}); err != nil {
		return err
	}

	return nil
}

func (dba dbAccess) toAccess() (*Access, error) {
	var err error
	var a = &Access{
		ExpiresAt: dba.ExpiresAt,
	}
	if a.Identity, err = id.ParsePublicKeyHex(dba.Identity); err != nil {
		return nil, err
	}
	if a.DataID, err = data.Parse(dba.DataID); err != nil {
		return nil, err
	}

	return a, nil
}

func toDbAccess(access *Access) dbAccess {
	return dbAccess{
		Identity:  access.Identity.String(),
		DataID:    access.DataID.String(),
		ExpiresAt: access.ExpiresAt,
	}
}
