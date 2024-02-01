package sets

import (
	"github.com/cryptopunkscc/astrald/mod/sets"
	"time"
)

type dbSet struct {
	ID                uint             `gorm:"primarykey"`
	Name              string           `gorm:"uniqueIndex"`
	Type              string           `gorm:"index"`
	Visible           bool             `gorm:"index;default:false;not null"`
	InclusionsAsSuper []dbSetInclusion `gorm:"foreignKey:SupersetID;OnDelete:CASCADE"`
	InclusionsAsSub   []dbSetInclusion `gorm:"foreignKey:SubsetID;OnDelete:CASCADE"`
	Description       string
	TrimmedAt         time.Time `gorm:"default:CURRENT_TIMESTAMP;NOT NULL"`
	CreatedAt         time.Time
}

func (dbSet) TableName() string { return sets.DBPrefix + "sets" }

func (mod *Module) dbFindSetByName(name string) (*dbSet, error) {
	var row dbSet
	var tx = mod.db.Where("name = ?", name).First(&row)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return &row, nil
}
