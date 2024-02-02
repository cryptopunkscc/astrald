package shares

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/shares"
	"time"
)

type dbRemoteDesc struct {
	Caller    id.Identity `gorm:"primaryKey"`
	Target    id.Identity `gorm:"primaryKey"`
	DataID    data.ID     `gorm:"primaryKey"`
	Desc      string
	CreatedAt time.Time `gorm:"index"`
}

func (dbRemoteDesc) TableName() string { return shares.DBPrefix + "remote_descs" }
