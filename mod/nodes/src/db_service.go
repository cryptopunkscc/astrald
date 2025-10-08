package nodes

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"time"
)

type dbService struct {
	ProviderID *astral.Identity `gorm:"primaryKey"`
	Name       string           `gorm:"primaryKey"`
	Priority   int              `gorm:"index"`
	ExpiresAt  time.Time        `gorm:"index"`
}

func (dbService) TableName() string {
	return nodes.DBPrefix + "services"
}
