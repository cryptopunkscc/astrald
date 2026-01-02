package services

import (
	"github.com/cryptopunkscc/astrald/astral"
)

type dbService struct {
	Name        astral.String8   `gorm:"primaryKey"`
	Identity    *astral.Identity `gorm:"primaryKey"`
	Composition *astral.Bundle   `gorm:"serializer:json"`
	Enabled     bool             `gorm:"index"`
	ExpiresAt   *astral.Time     `gorm:"index"` // Optional expiration time for cached service
}

func (dbService) TableName() string {
	return "services__services"
}
