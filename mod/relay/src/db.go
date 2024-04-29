package relay

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/relay"
	"github.com/cryptopunkscc/astrald/object"
	"gorm.io/gorm"
	"time"
)

type dbCert struct {
	DataID    object.ID      `gorm:"primaryKey"`
	TargetID  id.Identity    `gorm:"index"`
	RelayID   id.Identity    `gorm:"index"`
	Direction string         `gorm:"index"`
	ExpiresAt time.Time      `gorm:"index"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (dbCert) TableName() string {
	return relay.DBPrefix + "certs"
}
