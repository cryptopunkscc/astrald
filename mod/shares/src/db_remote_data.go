package shares

import (
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/mod/shares"
	"github.com/cryptopunkscc/astrald/object"
)

type dbRemoteData struct {
	Caller id.Identity `gorm:"primaryKey"`
	Target id.Identity `gorm:"primaryKey"`
	DataID object.ID   `gorm:"primaryKey"`
}

func (dbRemoteData) TableName() string {
	return shares.DBPrefix + "remote_data"
}
