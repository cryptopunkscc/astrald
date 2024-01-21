package index

import (
	"errors"
	_data "github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/index"
	"time"
)

type dbEntry struct {
	DataID uint `gorm:"primaryKey"`
	Data   dbData

	IndexID uint `gorm:"primaryKey"`
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

func (mod *Module) dbEntryCountByIndexID(indexID uint) (int, error) {
	var count int64
	var tx = mod.db.
		Model(&dbEntry{}).
		Where("index_id = ?", indexID).
		Count(&count)
	if tx.Error != nil {
		return -1, tx.Error
	}
	return int(count), nil
}

func (mod *Module) dbEntryFindByDataID(dataID uint) ([]dbEntry, error) {
	var rows []dbEntry
	var tx = mod.db.Where("data_id = ?", dataID).Preload("Index").Find(&rows)
	return rows, tx.Error
}

func (mod *Module) dbIndexSetAdded(indexID uint, dataID uint) (*dbEntry, error) {
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

	err = mod.db.Model(&row).Association("Data").Find(&row.Data)
	if err != nil {
		return nil, err
	}

	err = mod.db.Model(&row).Association("Index").Find(&row.Index)
	if err != nil {
		return nil, err
	}

	d, err := _data.Parse(row.Data.DataID)
	if err != nil {
		return nil, err
	}

	mod.events.Emit(index.EventEntryUpdate{
		IndexName: row.Index.Name,
		DataID:    d,
		Added:     true,
		UpdatedAt: row.UpdatedAt,
	})

	return row, nil
}

func (mod *Module) dbIndexAddDataID(indexes []uint, dataID uint) error {
	var added []uint
	for _, indexID := range indexes {
		_, err := mod.dbIndexSetAdded(indexID, dataID)
		if err != nil {
			continue
		}
		added = append(added, indexID)
	}

	if len(added) == 0 {
		return nil
	}

	var unionIDs []uint
	var tx = mod.db.
		Model(&dbUnion{}).
		Where("set_id in ?", added).
		Distinct("union_id").
		Find(&unionIDs)
	if tx.Error != nil {
		return tx.Error
	}

	return mod.dbIndexAddDataID(unionIDs, dataID)
}

func (mod *Module) dbIndexRemoveDataID(indexes []uint, dataID uint) error {
	var removed []uint
	for _, indexID := range indexes {
		_, err := mod.dbIndexSetRemoved(indexID, dataID)
		if err != nil {
			continue
		}
		removed = append(removed, indexID)
	}

	if len(removed) == 0 {
		return nil
	}

	// find direct supersets
	var unionIDs []uint
	var tx = mod.db.
		Model(&dbUnion{}).
		Where("set_id in ?", removed).
		Distinct("union_id").
		Find(&unionIDs)
	if tx.Error != nil {
		return tx.Error
	}

	// select supersets that no longer contain the dataid
	var selected []uint
	for _, unionID := range unionIDs {
		found, err := mod.dbUnionSubsetsContain(unionID, dataID)
		if err != nil {
			return err
		}
		if !found {
			selected = append(selected, unionID)
		}
	}

	return mod.dbIndexRemoveDataID(selected, dataID)
}

func (mod *Module) dbUnionSubsetsContain(unionID uint, dataID uint) (bool, error) {
	var subsets = mod.db.
		Model(&dbUnion{}).
		Where("union_id = ?", unionID).
		Distinct("set_id")

	var count int64

	var tx = mod.db.
		Model(&dbEntry{}).
		Where("data_id = ? and index_id in (?) and added = true", dataID, subsets).
		Count(&count)

	if tx.Error != nil {
		return false, tx.Error
	}

	return count > 0, nil
}

func (mod *Module) dbIndexSetRemoved(indexID uint, dataID uint) (*dbEntry, error) {
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
	if tx.Error != nil {
		return nil, tx.Error
	}

	err = mod.db.Model(&row).Association("Data").Find(&row.Data)
	if err != nil {
		return nil, err
	}

	err = mod.db.Model(&row).Association("Index").Find(&row.Index)
	if err != nil {
		return nil, err
	}

	d, err := _data.Parse(row.Data.DataID)
	if err != nil {
		return nil, err
	}

	mod.events.Emit(index.EventEntryUpdate{
		IndexName: row.Index.Name,
		DataID:    d,
		Added:     false,
		UpdatedAt: row.UpdatedAt,
	})

	return row, nil
}
