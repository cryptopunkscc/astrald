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
	case index.TypeSet, index.TypeUnion:
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

func (mod *Module) dbDeleteIndexByName(name string) error {
	indexRow, err := mod.dbFindIndexByName(name)
	if err != nil {
		return err
	}

	var tx = mod.db.Model(&dbIndex{}).Delete("id = ?", indexRow.ID)
	if tx.Error != nil {
		return tx.Error
	}

	return tx.Error
}
