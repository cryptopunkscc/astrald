package index

import (
	"context"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/index"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/tasks"
	"gorm.io/gorm"
	"time"
)

var _ index.Module = &Module{}

type Module struct {
	config Config
	node   node.Node
	log    *log.Logger
	assets assets.Store
	events events.Queue
	db     *gorm.DB
}

func (mod *Module) Run(ctx context.Context) error {
	return tasks.Group(
		&Service{Module: mod},
	).Run(ctx)
}

func (mod *Module) CreateIndex(name string, typ index.Type) (*index.Info, error) {
	row, err := mod.dbCreateIndex(name, string(typ))
	if err != nil {
		return nil, err
	}

	var info = &index.Info{
		Name:      row.Name,
		Type:      index.Type(row.Type),
		Size:      0,
		CreatedAt: row.CreatedAt,
	}

	mod.events.Emit(index.EventIndexCreated{Info: info})

	return info, nil
}

func (mod *Module) DeleteIndex(name string) error {
	var err = mod.dbDeleteIndex(name)
	if err != nil {
		return err
	}

	mod.events.Emit(index.EventIndexDeleted{Name: name})

	return nil
}

func (mod *Module) AddToSet(name string, dataID data.ID) error {
	indexRow, err := mod.dbFindIndexByName(name)
	if err != nil {
		return err
	}
	if index.Type(indexRow.Type) != index.TypeSet {
		return errors.New("index is not a set")
	}

	dataRow, err := mod.dbDataFindOrCreateByDataID(dataID.String())
	if err != nil {
		return fmt.Errorf("cannot get data row id: %w", err)
	}

	row, err := mod.dbEntryAddToIndex(indexRow.ID, dataRow.ID)
	if err == nil {
		mod.events.Emit(index.EventIndexEntryUpdate{
			IndexName: name,
			DataID:    dataID,
			Added:     true,
			UpdatedAt: row.UpdatedAt,
		})
	}

	return err
}

func (mod *Module) RemoveFromSet(name string, dataID data.ID) error {
	indexRow, err := mod.dbFindIndexByName(name)
	if err != nil {
		return err
	}
	if index.Type(indexRow.Type) != index.TypeSet {
		return errors.New("index is not a set")
	}

	dataRow, err := mod.dbDataFindOrCreateByDataID(dataID.String())
	if err != nil {
		return err
	}

	row, err := mod.dbEntryRemoveFromIndex(indexRow.ID, dataRow.ID)
	if err == nil {
		mod.events.Emit(index.EventIndexEntryUpdate{
			IndexName: name,
			DataID:    dataID,
			Added:     false,
			UpdatedAt: row.UpdatedAt,
		})
	}

	return err
}

func (mod *Module) IndexInfo(name string) (*index.Info, error) {
	indexRow, err := mod.dbFindIndexByName(name)
	if err != nil {
		return nil, err
	}

	var info = &index.Info{
		Name:      indexRow.Name,
		Type:      index.Type(indexRow.Type),
		Size:      0,
		CreatedAt: indexRow.CreatedAt,
	}

	var count int64

	var tx = mod.db.
		Model(&dbEntry{}).
		Where("index_id = ? and added = true", indexRow.ID).
		Count(&count)

	if tx.Error != nil {
		return nil, tx.Error
	}

	info.Size = int(count)

	return info, nil
}

func (mod *Module) Contains(name string, dataID data.ID) (bool, error) {
	indexRow, err := mod.dbFindIndexByName(name)
	if err != nil {
		return false, err
	}

	dataRow, err := mod.dbDataFindOrCreateByDataID(dataID.String())
	if err != nil {
		return false, err
	}

	_, err = mod.dbEntryFind(indexRow.ID, dataRow.ID)

	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return false, nil
	case err != nil:
		return false, err
	default:
		return true, nil
	}
}

func (mod *Module) Find(dataID data.ID) ([]string, error) {
	dataRow, err := mod.dbDataFindOrCreateByDataID(dataID.String())
	if err != nil {
		return nil, err
	}

	rows, err := mod.dbEntryFindByDataID(dataRow.ID)
	if err != nil {
		return nil, err
	}

	var list []string

	for _, row := range rows {
		if row.Added {
			list = append(list, row.Index.Name)
		}
	}

	return list, nil
}

func (mod *Module) AllIndexes() ([]index.Info, error) {
	var list []index.Info

	rows, err := mod.dbIndexFindAll()
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		var count int64

		var tx = mod.db.
			Model(&dbEntry{}).
			Where("index_id = ? and added = true", row.ID).
			Count(&count)

		if tx.Error != nil {
			mod.log.Errorv(2, "error getting entry count: %v", tx.Error)
			count = -1
		}

		list = append(list, index.Info{
			Name:      row.Name,
			Type:      index.Type(row.Type),
			Size:      int(count),
			CreatedAt: row.CreatedAt,
		})
	}

	return list, err
}

func (mod *Module) UpdatedSince(name string, since time.Time) ([]index.Entry, error) {
	indexRow, err := mod.dbFindIndexByName(name)
	if err != nil {
		return nil, err
	}

	rows, err := mod.dbEntryFindUpdatedSince(indexRow.ID, since)
	if err != nil {
		return nil, err
	}

	var updates []index.Entry

	for _, row := range rows {
		dataID, err := data.Parse(row.Data.DataID)
		if err != nil {
			return nil, err
		}

		updates = append(updates, index.Entry{
			DataID:    dataID,
			Added:     row.Added,
			UpdatedAt: row.UpdatedAt,
		})
	}

	return updates, nil
}
