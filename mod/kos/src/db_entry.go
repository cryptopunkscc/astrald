package kos

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/kos"
)

type dbEntry struct {
	Identity *astral.Identity `gorm:"primaryKey"`
	Key      string           `gorm:"primaryKey"`
	Type     string           `gorm:"index"`
	Payload  []byte
}

func (dbEntry) TableName() string { return kos.DBPrefix + "entries" }
