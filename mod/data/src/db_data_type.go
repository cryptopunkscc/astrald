package data

import (
	"time"
)

type dbDataType struct {
	DataID    string    `gorm:"primaryKey,index"`
	Header    string    `gorm:"index"`
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
