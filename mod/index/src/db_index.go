package index

import (
	"errors"
	"github.com/cryptopunkscc/astrald/mod/index"
	"time"
)

type dbIndex struct {
	ID        uint   `gorm:"primarykey"`
	Name      string `gorm:"uniqueIndex"`
	Type      string
	CreatedAt time.Time
}

func (dbIndex) TableName() string { return "indexes" }

func (mod *Module) dbIndexFindAll() ([]dbIndex, error) {
	var rows []dbIndex
	var tx = mod.db.Find(&rows)
	return rows, tx.Error
}

func (mod *Module) dbFindIndexByName(name string) (*dbIndex, error) {
	var row dbIndex
	var tx = mod.db.Where("name = ?", name).First(&row)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return &row, nil
}

func (mod *Module) dbFindIndexByNameAndType(name string, kind string) (*dbIndex, error) {
	var row dbIndex
	var tx = mod.db.Where("name = ? and type = ?", name, kind).First(&row)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return &row, nil
}

func (mod *Module) dbCreateIndex(name string, typ string) (*dbIndex, error) {
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}
	if typ == "" {
		typ = string(index.TypeSet)
	}

	switch index.Type(typ) {
	case index.TypeSet:
	default:
		return nil, errors.New("invalid index type")
	}

	var row = dbIndex{
		Name: name,
		Type: typ,
	}
	var tx = mod.db.Create(&row)

	return &row, tx.Error
}

func (mod *Module) dbDeleteIndex(name string) error {
	indexRow, err := mod.dbFindIndexByName(name)
	if err != nil {
		return err
	}

	//// find all entries in the index
	//var entries []dbIndexEntry
	//var tx = mod.db.
	//	Where("index_id = ? and added = true", indexRow.ID).
	//	Preload("Data").
	//	Find(&entries)
	//if tx.Error != nil {
	//	return tx.Error
	//}
	//
	//// mark all entries as removed
	//for _, entry := range entries {
	//	dataID, err := data.Parse(entry.Data.DataID)
	//	if err != nil {
	//		mod.log.Errorv(2, "parse '%v' error: %v", entry.Data.DataID, err)
	//		continue
	//	}
	//
	//	row, err := mod.dbEntryRemoveFromIndex(indexRow.ID, entry.DataID)
	//	if err != nil {
	//		mod.log.Errorv(2, "remove %v from %v: %v", entry.DataID, indexRow.ID, err)
	//		continue
	//	}
	//	mod.events.Emit(storage.EventIndexEntryUpdate{
	//		IndexName: name,
	//		DataID:    dataID,
	//		Added:     false,
	//		UpdatedAt: row.UpdatedAt,
	//	})
	//}
	//
	//tx = mod.db.Model(&dbIndexEntry{}).Delete("index_id = ?", indexRow.ID)
	//if tx.Error != nil {
	//	return tx.Error
	//}

	var tx = mod.db.Model(&dbIndex{}).Delete("id = ?", indexRow.ID)
	if tx.Error != nil {
		return tx.Error
	}

	return tx.Error
}
