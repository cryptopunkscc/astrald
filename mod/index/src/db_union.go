package index

type dbUnion struct {
	UnionID uint `gorm:"primaryKey"`
	Union   *dbIndex
	SetID   uint `gorm:"primaryKey"`
	Set     *dbIndex
}

func (dbUnion) TableName() string { return "unions" }

func (mod *Module) dbUnionCreate(unionID uint, setID uint) error {
	var tx = mod.db.Create(&dbUnion{
		SetID:   setID,
		UnionID: unionID,
	})

	return tx.Error
}

func (mod *Module) dbUnionFindBySetID(setID uint) ([]dbUnion, error) {
	var rows []dbUnion

	var tx = mod.db.
		Where("set_id = ?", setID).
		Preload("Union").
		Find(&rows)

	return rows, tx.Error
}

func (mod *Module) dbUnionFindByUnionID(unionID uint) ([]dbUnion, error) {
	var rows []dbUnion

	var tx = mod.db.
		Where("union_id = ?", unionID).
		Preload("Set").
		Find(&rows)

	return rows, tx.Error
}

func (mod *Module) dbUnionContains(unionID uint, dataID uint) (bool, error) {
	unions, err := mod.dbUnionFindByUnionID(unionID)
	if err != nil {
		return false, err
	}

	for _, union := range unions {
		entry, err := mod.dbEntryFind(union.SetID, dataID)
		if (err == nil) && entry.Added {
			return true, nil
		}
	}

	return false, nil
}
