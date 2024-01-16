package index

import (
	"errors"
	"time"
)

type dbEntry struct {
	DataID uint
	Data   dbData

	IndexID uint
	Index   dbIndex

	Added     bool `gorm:"default:true;not null"`
	UpdatedAt time.Time
}

func (dbEntry) TableName() string { return "entries" }

func (mod *Module) dbEntryCreate(indexID uint, dataID uint) (*dbEntry, error) {
	var row = dbEntry{
		DataID:  dataID,
		IndexID: indexID,
	}

	var tx = mod.db.Create(&row)

	return &row, tx.Error
}

func (mod *Module) dbEntryFind(indexID uint, dataID uint) (*dbEntry, error) {
	var row dbEntry
	var tx = mod.db.Where("index_id = ? and data_id = ?", indexID, dataID).First(&row)
	return &row, tx.Error
}

func (mod *Module) dbEntryFindByIndexID(indexID uint) ([]dbEntry, error) {
	var rows []dbEntry
	var tx = mod.db.Where("index_id = ?", indexID).Preload("Data").Find(&rows)
	return rows, tx.Error
}

func (mod *Module) dbEntryFindByDataID(dataID uint) ([]dbEntry, error) {
	var rows []dbEntry
	var tx = mod.db.Where("data_id = ?", dataID).Preload("Index").Find(&rows)
	return rows, tx.Error
}

func (mod *Module) dbEntryAddToIndex(indexID uint, dataID uint) (*dbEntry, error) {
	row, err := mod.dbEntryFind(indexID, dataID)
	if err != nil {
		row, err = mod.dbEntryCreate(indexID, dataID)
		if err != nil {
			return nil, err
		}
	} else {
		if row.Added {
			return nil, errors.New("already added")
		}
		row.Added = true
		var tx = mod.db.
			Where("data_id = ? and index_id = ?", dataID, indexID).
			Save(&row)
		if tx.Error != nil {
			return nil, tx.Error
		}
	}

	return row, nil
}

func (mod *Module) dbEntryRemoveFromIndex(indexID uint, dataID uint) (*dbEntry, error) {
	row, err := mod.dbEntryFind(indexID, dataID)
	if err != nil {
		return nil, err
	}
	if !row.Added {
		return nil, errors.New("already removed")
	}

	row.Added = false
	var tx = mod.db.
		Where("data_id = ? and index_id = ?", dataID, indexID).
		Save(&row)

	return row, tx.Error
}

func (mod *Module) dbEntryFindUpdatedSince(indexID uint, since time.Time) ([]dbEntry, error) {
	var rows []dbEntry

	query := mod.db.
		Where("index_id = ?", indexID).
		Order("updated_at asc").
		Preload("Data")

	if !since.IsZero() {
		query = query.Where("updated_at > ?", since)
	}

	var tx = query.Find(&rows)

	return rows, tx.Error
}
