package data

import (
	"context"
	"github.com/cryptopunkscc/astrald/cslq"
	_data "github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/data"
	"github.com/cryptopunkscc/astrald/mod/storage"
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

// SubscribeType returns a channel that will be populated with all data entries since the provided timestamp and
// subscribed to any new items until context is done. If typ is empty, all data entries will be passed regardless
// of the type.
func (mod *Module) SubscribeType(ctx context.Context, typ string, since time.Time) <-chan data.TypeInfo {
	if since.After(time.Now()) {
		return nil
	}

	var ch = make(chan data.TypeInfo)
	var subscription = mod.events.Subscribe(ctx)

	go func() {
		defer close(ch)

		// catch up with existing entries
		list, err := mod.FindByType(typ, since)
		if err != nil {
			return
		}
		for _, item := range list {
			select {
			case ch <- item:
			case <-ctx.Done():
				return
			}
		}

		// subscribe to new items
		for event := range subscription {
			e, ok := event.(data.EventDataIdentified)
			if !ok {
				continue
			}
			if typ != "" && e.Type != typ {
				continue
			}
			ch <- data.TypeInfo(e)
		}
	}()

	return ch
}

// FindByType returns all data items indexed since time ts. If t is not empty, only items of type t will
// be returned.
func (mod *Module) FindByType(t string, ts time.Time) ([]data.TypeInfo, error) {
	var list []data.TypeInfo
	var rows []*dbDataType

	// filter by type if provided
	var q = mod.db
	if t != "" {
		q = q.Where("type = ?", t)
	}

	// fetch rows
	var tx = q.Where("indexed_at > ?", ts).Find(&rows)
	if tx.Error != nil {
		return nil, tx.Error
	}

	for _, row := range rows {
		dataID, err := _data.Parse(row.DataID)
		if err != nil {
			continue
		}

		list = append(list, data.TypeInfo{
			DataID:    dataID,
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
			DataID:    dataID,
			IndexedAt: row.IndexedAt,
			Header:    row.Header,
			Type:      row.Type,
		})
	}

	return list, nil
}

func (mod *Module) OpenADC0(dataID _data.ID) (string, storage.DataReader, error) {
	reader, err := mod.storage.Data().Read(dataID, nil)
	if err != nil {
		return "", nil, err
	}

	var header data.ADC0Header
	err = cslq.Decode(reader, "v", &header)
	if err != nil {
		return "", nil, err
	}

	return string(header), reader, nil
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
