package nodes

import (
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"time"
)

type dbEndpoint struct {
	Identity  id.Identity `gorm:"primaryKey"`
	Network   string      `gorm:"primaryKey"`
	Address   string      `gorm:"primaryKey"`
	CreatedAt time.Time
}

func (dbEndpoint) TableName() string {
	return nodes.DBPrefix + "endpoints"
}
