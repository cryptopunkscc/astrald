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

func (mod *Module) dbUnionDelete(unionID uint, setID uint) error {
	var tx = mod.db.Delete(&dbUnion{
		SetID:   setID,
		UnionID: unionID,
	})

	return tx.Error
}

func (mod *Module) dbUnionFindByUnionID(unionID uint) ([]dbUnion, error) {
	var rows []dbUnion

	var tx = mod.db.
		Where("union_id = ?", unionID).
		Preload("Set").
		Find(&rows)

	return rows, tx.Error
}
