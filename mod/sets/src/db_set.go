package sets

import (
	"github.com/cryptopunkscc/astrald/mod/sets"
	"time"
)

type dbSet struct {
	ID        uint      `gorm:"primarykey"`
	Name      string    `gorm:"uniqueIndex"`
	TrimmedAt time.Time `gorm:"default:CURRENT_TIMESTAMP;NOT NULL"`
	CreatedAt time.Time `gorm:"index"`
}

func (dbSet) TableName() string { return sets.DBPrefix + "sets" }
