package shares

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/shares"
)

type dbRemoteData struct {
	Caller id.Identity `gorm:"primaryKey"`
	Target id.Identity `gorm:"primaryKey"`
	DataID data.ID     `gorm:"primaryKey"`
}

func (dbRemoteData) TableName() string {
	return shares.DBPrefix + "remote_data"
}
