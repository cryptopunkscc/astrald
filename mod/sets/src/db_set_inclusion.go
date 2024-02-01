package sets

import (
	"github.com/cryptopunkscc/astrald/mod/sets"
)

type dbSetInclusion struct {
	SupersetID uint `gorm:"primaryKey"`
	Superset   *dbSet
	SubsetID   uint `gorm:"primaryKey"`
	Subset     *dbSet
}

func (dbSetInclusion) TableName() string { return sets.DBPrefix + "set_inclusions" }
