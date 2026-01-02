package services

import (
	"github.com/cryptopunkscc/astrald/astral"
)

type dbService struct {
	ID          uint             `gorm:"primaryKey;autoIncrement"` // FIXME: remove this field
	Name        astral.String8   `gorm:"index"`
	Identity    *astral.Identity `gorm:"index"`
	Composition *astral.Bundle   `gorm:"serializer:json"`
	Enabled     bool             `gorm:"index"`
	CreatedAt   astral.Time      `gorm:"autoCreateTime"`
	ExpiresAt   astral.Time      `gorm:"index"` // When this record expires (earlier date = expired/invalidated)
}

func (dbService) TableName() string {
	return "services__services"
}
