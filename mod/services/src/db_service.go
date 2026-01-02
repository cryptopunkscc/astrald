package services

import (
	"github.com/cryptopunkscc/astrald/astral"
)

type dbService struct {
	Name        astral.String8   `gorm:"index"`
	Identity    *astral.Identity `gorm:"index"`
	Composition *astral.Bundle   `gorm:"serializer:json"`
	CreatedAt   astral.Time      `gorm:"autoCreateTime"`
	ExpiresAt   astral.Time      `gorm:"index"`
}

func (dbService) TableName() string {
	return "services__services"
}
