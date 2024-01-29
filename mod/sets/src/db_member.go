package sets

import (
	"errors"
	"github.com/cryptopunkscc/astrald/mod/sets"
	"time"
)

type dbMember struct {
	DataID uint `gorm:"primaryKey"`
	Data   dbData

	SetID uint `gorm:"primaryKey"`
	Set   dbSet

	Added     bool `gorm:"default:true;not null"`
	UpdatedAt time.Time
}

func (dbMember) TableName() string { return sets.DBPrefix + "members" }

func (mod *Module) dbMemberCreate(setID uint, dataID uint) (*dbMember, error) {
	var row = dbMember{
		DataID: dataID,
		SetID:  setID,
	}

	var tx = mod.db.Create(&row)

	return &row, tx.Error
}

func (mod *Module) dbMemberFind(setID uint, dataID uint) (*dbMember, error) {
	var row dbMember
	var tx = mod.db.Where("set_id = ? and data_id = ?", setID, dataID).First(&row)
	return &row, tx.Error
}

func (mod *Module) dbMemberFindBySetID(setID uint) ([]dbMember, error) {
	var rows []dbMember
	var tx = mod.db.Where("set_id = ?", setID).Preload("Data").Find(&rows)
	return rows, tx.Error
}

func (mod *Module) dbMemberFindUpdatedBetween(setID uint, since time.Time, until time.Time) ([]dbMember, error) {
	var rows []dbMember

	query := mod.db.
		Where("set_id = ?", setID).
		Order("updated_at asc").
		Preload("Data")

	if !since.IsZero() {
		query = query.Where("updated_at > ?", since)
	}
	if !until.IsZero() {
		query = query.Where("updated_at < ?", until)
	}

	var tx = query.Find(&rows)

	return rows, tx.Error
}

func (mod *Module) dbMemberCountBySetID(setID uint) (int, error) {
	var count int64
	var tx = mod.db.
		Model(&dbMember{}).
		Where("set_id = ?", setID).
		Count(&count)
	if tx.Error != nil {
		return -1, tx.Error
	}
	return int(count), nil
}

func (mod *Module) dbMemberFindByDataID(dataID uint) ([]dbMember, error) {
	var rows []dbMember
	var tx = mod.db.Where("data_id = ?", dataID).Preload("Set").Find(&rows)
	return rows, tx.Error
}

func (mod *Module) dbSetMarkAsAdded(setID uint, dataID uint) (*dbMember, error) {
	row, err := mod.dbMemberFind(setID, dataID)
	if err != nil {
		row, err = mod.dbMemberCreate(setID, dataID)
		if err != nil {
			return nil, err
		}
	} else {
		if row.Added {
			return nil, errors.New("already added")
		}
		row.Added = true
		var tx = mod.db.
			Where("data_id = ? and set_id = ?", dataID, setID).
			Save(&row)
		if tx.Error != nil {
			return nil, tx.Error
		}
	}

	err = mod.db.Model(&row).Association("Data").Find(&row.Data)
	if err != nil {
		return nil, err
	}

	err = mod.db.Model(&row).Association("Set").Find(&row.Set)
	if err != nil {
		return nil, err
	}

	mod.events.Emit(sets.EventEntryUpdate{
		SetName:   row.Set.Name,
		DataID:    row.Data.DataID,
		Added:     true,
		UpdatedAt: row.UpdatedAt,
	})

	return row, nil
}

func (mod *Module) dbSetAddDataID(setIDs []uint, dataID uint) error {
	var added []uint
	for _, setID := range setIDs {
		_, err := mod.dbSetMarkAsAdded(setID, dataID)
		if err != nil {
			continue
		}
		added = append(added, setID)
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

	return mod.dbSetAddDataID(unionIDs, dataID)
}

func (mod *Module) dbSetRemoveDataID(setIDs []uint, dataID uint) error {
	var removed []uint
	for _, setID := range setIDs {
		_, err := mod.dbSetMarkAsRemoved(setID, dataID)
		if err != nil {
			continue
		}
		removed = append(removed, setID)
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

	return mod.dbSetRemoveDataID(selected, dataID)
}

func (mod *Module) dbUnionSubsetsContain(unionID uint, dataID uint) (bool, error) {
	var subsets = mod.db.
		Model(&dbUnion{}).
		Where("union_id = ?", unionID).
		Distinct("set_id")

	var count int64

	var tx = mod.db.
		Model(&dbMember{}).
		Where("data_id = ? and set_id in (?) and added = true", dataID, subsets).
		Count(&count)

	if tx.Error != nil {
		return false, tx.Error
	}

	return count > 0, nil
}

func (mod *Module) dbSetMarkAsRemoved(setID uint, dataID uint) (*dbMember, error) {
	row, err := mod.dbMemberFind(setID, dataID)
	if err != nil {
		return nil, err
	}
	if !row.Added {
		return nil, errors.New("already removed")
	}

	row.Added = false
	var tx = mod.db.
		Where("data_id = ? and set_id = ?", dataID, setID).
		Save(&row)
	if tx.Error != nil {
		return nil, tx.Error
	}

	err = mod.db.Model(&row).Association("Data").Find(&row.Data)
	if err != nil {
		return nil, err
	}

	err = mod.db.Model(&row).Association("Set").Find(&row.Set)
	if err != nil {
		return nil, err
	}

	mod.events.Emit(sets.EventEntryUpdate{
		SetName:   row.Set.Name,
		DataID:    row.Data.DataID,
		Added:     false,
		UpdatedAt: row.UpdatedAt,
	})

	return row, nil
}
