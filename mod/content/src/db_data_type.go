package content

import (
	"github.com/cryptopunkscc/astrald/data"
	"time"
)

type dbDataType struct {
	DataID    data.ID   `gorm:"primaryKey,index"`
	Method    string    `gorm:"index"`
	Type      string    `gorm:"index"`
	IndexedAt time.Time `gorm:"index"`
}

func (dbDataType) TableName() string {
	return "data_types"
}

func (mod *Module) dbDataTypeFindByDataID(dataID string) (*dbDataType, error) {
	var row dbDataType
	var tx = mod.db.Where("data_id = ?", dataID).First(&row)
	return &row, tx.Error
}
