package data

import (
	"github.com/cryptopunkscc/astrald/cslq"
	_data "github.com/cryptopunkscc/astrald/data"
	data "github.com/cryptopunkscc/astrald/mod/data/api"
	"io"
	"time"
)

type dbDataType struct {
	DataID    string    `gorm:"primaryKey,index"`
	Header    string    `gorm:"index"`
	Type      string    `gorm:"index"`
	IndexedAt time.Time `gorm:"index"`
}

func (dbDataType) TableName() string {
	return "data_types"
}

func (mod *Module) FindByType(t string, ts time.Time) ([]data.TypeInfo, error) {
	var list []data.TypeInfo

	var rows []*dbDataType
	tx := mod.db.Where("type = ? and indexed_at > ?", t, ts).Find(&rows)
	if tx.Error != nil {
		return nil, tx.Error
	}

	for _, row := range rows {
		dataID, err := _data.Parse(row.DataID)
		if err != nil {
			continue
		}

		list = append(list, data.TypeInfo{
			ID:        dataID,
			IndexedAt: row.IndexedAt,
			Type:      row.Type,
		})
	}

	return list, nil
}

func (mod *Module) All(ts time.Time) ([]data.TypeInfo, error) {
	var list []data.TypeInfo

	var rows []*dbDataType
	tx := mod.db.Where("indexed_at > ?", ts).Find(&rows)
	if tx.Error != nil {
		return nil, tx.Error
	}

	for _, row := range rows {
		dataID, err := _data.Parse(row.DataID)
		if err != nil {
			continue
		}

		list = append(list, data.TypeInfo{
			ID:        dataID,
			IndexedAt: row.IndexedAt,
			Header:    row.Header,
			Type:      row.Type,
		})
	}

	return list, nil
}

func (mod *Module) OpenADC0(dataID _data.ID) (*data.ADC0Header, io.ReadCloser, error) {
	reader, err := mod.storage.Data().Read(dataID, nil)
	if err != nil {
		return nil, nil, err
	}

	var header data.ADC0Header
	err = cslq.Decode(reader, "v", &header)
	if err != nil {
		return nil, nil, err
	}

	return &header, reader, nil
}

func (mod *Module) mostRecentContainer() (*dbDataType, error) {
	var row dbDataType

	tx := mod.db.Order("added_at desc").Limit(1).First(&row)

	if tx.Error != nil {
		return nil, tx.Error
	}

	return &row, nil
}

func (mod *Module) findByDataID(dataID _data.ID) (*dbDataType, error) {
	var row dbDataType

	tx := mod.db.Where("data_id = ?", dataID.String()).First(&row)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return &row, nil
}
