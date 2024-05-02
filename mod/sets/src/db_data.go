package sets

import (
	"github.com/cryptopunkscc/astrald/mod/sets"
	"github.com/cryptopunkscc/astrald/object"
	"time"
)

type dbData struct {
	ID        uint      `gorm:"primarykey"`
	DataID    object.ID `gorm:"uniqueIndex"`
	CreatedAt time.Time `gorm:"index"`
}

func (dbData) TableName() string { return sets.DBPrefix + "data" }

func (mod *Module) dbDataFindByObjectID(objectID object.ID) (*dbData, error) {
	var row dbData
	var tx = mod.db.Where("data_id = ?", objectID).First(&row)
	return &row, tx.Error
}

func (mod *Module) dbDataCreate(objectID object.ID) (*dbData, error) {
	var row = dbData{DataID: objectID}
	var tx = mod.db.Create(&row)
	return &row, tx.Error
}

func (mod *Module) dbDataFindOrCreateByObjectID(objectID object.ID) (*dbData, error) {
	if row, err := mod.dbDataFindByObjectID(objectID); err == nil {
		return row, nil
	}

	return mod.dbDataCreate(objectID)
}
