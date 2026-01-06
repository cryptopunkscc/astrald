package services

import (
	"github.com/cryptopunkscc/astrald/astral"
)

// dbService represents discovered service of a specific identity cached in the database
type dbService struct {
	Name        astral.String8   `gorm:"index"`
	Identity    *astral.Identity `gorm:"index"`
	Composition *astral.Bundle   `gorm:"serializer:json"`
	CreatedAt   astral.Time      `gorm:"autoCreateTime"`
}

func (dbService) TableName() string {
	return "services__services"
}
