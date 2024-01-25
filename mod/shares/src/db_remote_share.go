package shares

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"time"
)

type dbRemoteShare struct {
	Caller     id.Identity `gorm:"primaryKey"`
	Target     id.Identity `gorm:"primaryKey"`
	LastUpdate time.Time
}

func (dbRemoteShare) TableName() string {
	return "remote_shares"
}
