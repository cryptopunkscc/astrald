package sets

import (
	"errors"
	"github.com/cryptopunkscc/astrald/mod/sets"
	"time"
)

type dbSet struct {
	ID          uint   `gorm:"primarykey"`
	Name        string `gorm:"uniqueIndex"`
	Type        string `gorm:"index"`
	Visible     bool   `gorm:"index;default:false;not null"`
	Description string
	CreatedAt   time.Time
}

func (dbSet) TableName() string { return sets.DBPrefix + "sets" }

func (mod *Module) dbSetFindAll() ([]dbSet, error) {
	var rows []dbSet
	var tx = mod.db.Find(&rows)
	return rows, tx.Error
}

func (mod *Module) dbFindSetByName(name string) (*dbSet, error) {
	var row dbSet
	var tx = mod.db.Where("name = ?", name).First(&row)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return &row, nil
}

func (mod *Module) dbSetUpdateVisible(name string, visible bool) error {
	return mod.db.
		Model(&dbSet{}).
		Where("name = ?", name).
		Update("visible", visible).
		Error
}

func (mod *Module) dbSetSetDescription(name string, desc string) error {
	return mod.db.
		Model(&dbSet{}).
		Where("name = ?", name).
		Update("description", desc).
		Error
}

func (mod *Module) dbFindSetByNameAndType(name string, kind string) (*dbSet, error) {
	var row dbSet
	var tx = mod.db.Where("name = ? and type = ?", name, kind).First(&row)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return &row, nil
}

func (mod *Module) dbCreateSet(name string, typ string) (*dbSet, error) {
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}
	if typ == "" {
		typ = string(sets.TypeSet)
	}

	switch sets.Type(typ) {
	case sets.TypeSet, sets.TypeUnion:
	default:
		return nil, errors.New("invalid set type")
	}

	var row = dbSet{
		Name: name,
		Type: typ,
	}
	var tx = mod.db.Create(&row)

	return &row, tx.Error
}

func (mod *Module) dbDeleteSetByName(name string) error {
	setRow, err := mod.dbFindSetByName(name)
	if err != nil {
		return err
	}

	var tx = mod.db.Model(&dbSet{}).Delete("id = ?", setRow.ID)
	if tx.Error != nil {
		return tx.Error
	}

	return tx.Error
}
