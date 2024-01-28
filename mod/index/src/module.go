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
	assets assets.Assets
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
	indexRow, err := mod.dbFindIndexByName(name)
	if err != nil {
		return err
	}

	// remove all entries from the index
	entryRows, err := mod.dbEntryFindUpdatedBetween(indexRow.ID, time.Time{}, time.Time{})
	if err != nil {
		return err
	}

	for _, row := range entryRows {
		mod.removeFromIndex(indexRow, row.Data.DataID)
	}

	var tx = mod.db.Model(&dbUnion{}).Delete("set_id = ?", indexRow.ID)
	if tx.Error != nil {
		return tx.Error
	}

	err = mod.dbDeleteIndexByName(name)
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

	return mod.addToIndex(indexRow, dataID)
}

func (mod *Module) addToIndex(indexRow *dbIndex, dataID data.ID) error {
	dataRow, err := mod.dbDataFindOrCreateByDataID(dataID)
	if err != nil {
		return err
	}

	return mod.dbIndexAddDataID([]uint{indexRow.ID}, dataRow.ID)
}

func (mod *Module) RemoveFromSet(name string, dataID data.ID) error {
	indexRow, err := mod.dbFindIndexByName(name)
	if err != nil {
		return err
	}
	if index.Type(indexRow.Type) != index.TypeSet {
		return errors.New("index is not a set")
	}

	return mod.removeFromIndex(indexRow, dataID)
}

func (mod *Module) removeFromIndex(indexRow *dbIndex, dataID data.ID) error {
	dataRow, err := mod.dbDataFindOrCreateByDataID(dataID)
	if err != nil {
		return err
	}

	return mod.dbIndexRemoveDataID([]uint{indexRow.ID}, dataRow.ID)
}

func (mod *Module) IndexInfo(name string) (*index.Info, error) {
	indexRow, err := mod.dbFindIndexByName(name)
	if err != nil {
		return nil, err
	}

	var info = &index.Info{
		Name:        indexRow.Name,
		Type:        index.Type(indexRow.Type),
		Size:        0,
		Visible:     indexRow.Visible,
		Description: indexRow.Description,
		CreatedAt:   indexRow.CreatedAt,
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, index.ErrIndexNotFound
		}
		return false, fmt.Errorf("cannot read index %s: %w", name, err)
	}

	dataRow, err := mod.dbDataFindOrCreateByDataID(dataID)
	if err != nil {
		return false, err
	}

	entry, err := mod.dbEntryFind(indexRow.ID, dataRow.ID)

	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return false, nil
	case err != nil:
		return false, err
	default:
		return entry.Added, nil
	}
}

func (mod *Module) Find(dataID data.ID) ([]string, error) {
	dataRow, err := mod.dbDataFindOrCreateByDataID(dataID)
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
			Name:        row.Name,
			Type:        index.Type(row.Type),
			Size:        int(count),
			Visible:     row.Visible,
			Description: row.Description,
			CreatedAt:   row.CreatedAt,
		})
	}

	return list, err
}

func (mod *Module) UpdatedBetween(name string, since time.Time, until time.Time) ([]index.Entry, error) {
	indexRow, err := mod.dbFindIndexByName(name)
	if err != nil {
		return nil, err
	}

	rows, err := mod.dbEntryFindUpdatedBetween(indexRow.ID, since, until)
	if err != nil {
		return nil, err
	}

	var updates []index.Entry

	for _, row := range rows {
		updates = append(updates, index.Entry{
			DataID:    row.Data.DataID,
			Added:     row.Added,
			UpdatedAt: row.UpdatedAt,
		})
	}

	return updates, nil
}

func (mod *Module) Scan(name string, opts *index.ScanOpts) ([]*index.Entry, error) {
	indexRow, err := mod.dbFindIndexByName(name)
	if err != nil {
		return nil, err
	}

	if opts == nil {
		opts = &index.ScanOpts{}
	}

	var rows []dbEntry
	var q = mod.db.
		Where("index_id = ?", indexRow.ID).
		Order("updated_at").
		Preload("Data")

	if !opts.UpdatedAfter.IsZero() {
		q = q.Where("updated_at > ?", opts.UpdatedAfter)
	}
	if !opts.UpdatedBefore.IsZero() {
		q = q.Where("updated_at < ?", opts.UpdatedBefore)
	}
	if !opts.IncludeRemoved {
		q = q.Where("added = true")
	}

	err = q.Find(&rows).Error
	if err != nil {
		return nil, err
	}

	var entries []*index.Entry

	for _, row := range rows {
		entries = append(entries, &index.Entry{
			DataID:    row.Data.DataID,
			Added:     row.Added,
			UpdatedAt: row.UpdatedAt,
		})
	}

	return entries, nil
}

func (mod *Module) AddToUnion(union string, index string) error {
	unionRow, err := mod.dbFindIndexByName(union)
	if err != nil {
		return err
	}

	setRow, err := mod.dbFindIndexByName(index)
	if err != nil {
		return err
	}

	err = mod.dbUnionCreate(unionRow.ID, setRow.ID)
	if err != nil {
		return err
	}

	entries, err := mod.dbEntryFindUpdatedBetween(setRow.ID, time.Time{}, time.Time{})
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.Added {
			continue
		}

		err = mod.addToIndex(unionRow, entry.Data.DataID)
		if err != nil {
			mod.log.Errorv(2,
				"error adding %v to union %v: %v",
				entry.Data.DataID,
				unionRow.Name,
				err,
			)
		}
	}

	return nil
}

func (mod *Module) RemoveFromUnion(union string, index string) error {
	unionRow, err := mod.dbFindIndexByName(union)
	if err != nil {
		return err
	}

	indexRow, err := mod.dbFindIndexByName(index)
	if err != nil {
		return err
	}

	err = mod.dbUnionDelete(unionRow.ID, indexRow.ID)
	if err != nil {
		return err
	}

	var dataIDs []uint

	err = mod.db.
		Model(&dbEntry{}).
		Where("index_id = ? and added = true", indexRow.ID).
		Select("data_id").
		Find(&dataIDs).Error
	if err != nil {
		return err
	}

	for _, dataID := range dataIDs {
		found, err := mod.dbUnionSubsetsContain(unionRow.ID, dataID)
		if err != nil {
			return err
		}
		if !found {
			mod.dbIndexRemoveDataID([]uint{unionRow.ID}, dataID)
		}
	}

	return nil
}

func (mod *Module) GetEntry(name string, dataID data.ID) (*index.Entry, error) {
	indexRow, err := mod.dbFindIndexByName(name)
	if err != nil {
		return nil, err
	}

	dataRow, err := mod.dbDataFindOrCreateByDataID(dataID)
	if err != nil {
		return nil, err
	}

	row, err := mod.dbEntryFind(indexRow.ID, dataRow.ID)
	if err != nil {
		return nil, err
	}

	return &index.Entry{
		DataID:    dataID,
		Added:     row.Added,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

func (mod *Module) SetVisible(name string, visible bool) error {
	return mod.dbIndexSetVisible(name, visible)
}

func (mod *Module) SetDescription(name string, desc string) error {
	return mod.dbIndexSetDescription(name, desc)
}
