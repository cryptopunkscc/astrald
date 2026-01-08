package services

import (
	"github.com/cryptopunkscc/astrald/astral"
)

// dbService represents discovered service of a specific identity cached in the database
type dbService struct {
	Name        astral.String8   `gorm:"uniqueIndex:idx_db_service_name_identity"`
	Identity    *astral.Identity `gorm:"uniqueIndex:idx_db_service_name_identity"`
	Composition *astral.Bundle   `gorm:"serializer:json"`
	CreatedAt   astral.Time      `gorm:"autoCreateTime"`
}

func (dbService) TableName() string {
	return "services__services"
}
