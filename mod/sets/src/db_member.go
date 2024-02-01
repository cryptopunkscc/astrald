package sets

import (
	"github.com/cryptopunkscc/astrald/mod/sets"
	"time"
)

type dbMember struct {
	DataID uint `gorm:"primaryKey"`
	Data   *dbData

	SetID uint `gorm:"primaryKey"`
	Set   *dbSet

	Removed   bool `gorm:"default:false;not null"`
	UpdatedAt time.Time
}

func (dbMember) TableName() string { return sets.DBPrefix + "members" }
