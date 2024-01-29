package sets

import "github.com/cryptopunkscc/astrald/mod/sets"

type dbUnion struct {
	UnionID uint `gorm:"primaryKey"`
	Union   *dbSet
	SetID   uint `gorm:"primaryKey"`
	Set     *dbSet
}

func (dbUnion) TableName() string { return sets.DBPrefix + "unions" }

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
