package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/user"
	"github.com/cryptopunkscc/astrald/object"
)

type dbAsset struct {
	Nonce    astral.Nonce `gorm:"primaryKey"`
	Removed  bool         `gorm:"index"`
	ObjectID *object.ID   `gorm:"index"`
	Height   uint64       `gorm:"uniqueIndex"`
}

func (dbAsset) TableName() string { return user.DBPrefix + "assets" }
