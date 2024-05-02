package shares

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/shares"
	"github.com/cryptopunkscc/astrald/object"
	"time"
)

type dbRemoteDesc struct {
	Caller    id.Identity `gorm:"primaryKey"`
	Target    id.Identity `gorm:"primaryKey"`
	DataID    object.ID   `gorm:"primaryKey"`
	Desc      string
	CreatedAt time.Time `gorm:"index"`
}

func (dbRemoteDesc) TableName() string { return shares.DBPrefix + "remote_descs" }
