package shares

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/shares"
	"time"
)

type dbRemoteShare struct {
	Caller     id.Identity `gorm:"primaryKey"`
	Target     id.Identity `gorm:"primaryKey"`
	SetName    string      `gorm:"uniqueIndex;not null;default:null"`
	LastUpdate time.Time
}

func (dbRemoteShare) TableName() string {
	return shares.DBPrefix + "remote_shares"
}
