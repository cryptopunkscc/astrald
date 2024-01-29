package sets

import (
	"context"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/sets"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/tasks"
	"gorm.io/gorm"
	"time"
)

var _ sets.Module = &Module{}

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

func (mod *Module) CreateSet(name string, typ sets.Type) (*sets.Info, error) {
	row, err := mod.dbCreateSet(name, string(typ))
	if err != nil {
		return nil, err
	}

	var info = &sets.Info{
		Name:      row.Name,
		Type:      sets.Type(row.Type),
		Size:      0,
		CreatedAt: row.CreatedAt,
	}

	mod.events.Emit(sets.EventSetCreated{Info: info})

	return info, nil
}

func (mod *Module) DeleteSet(name string) error {
	setRow, err := mod.dbFindSetByName(name)
	if err != nil {
		return err
	}

	// remove all entries from the set
	entryRows, err := mod.dbMemberFindUpdatedBetween(setRow.ID, time.Time{}, time.Time{})
	if err != nil {
		return err
	}

	for _, row := range entryRows {
		mod.removeFromSet(setRow, row.Data.DataID)
	}

	var tx = mod.db.Model(&dbUnion{}).Delete("set_id = ?", setRow.ID)
	if tx.Error != nil {
		return tx.Error
	}

	err = mod.dbDeleteSetByName(name)
	if err != nil {
		return err
	}

	mod.events.Emit(sets.EventSetDeleted{Name: name})

	return nil
}

func (mod *Module) AddToSet(name string, dataID data.ID) error {
	setRow, err := mod.dbFindSetByName(name)
	if err != nil {
		return err
	}
	if sets.Type(setRow.Type) != sets.TypeSet {
		return sets.ErrInvalidSetType
	}

	return mod.addToSet(setRow, dataID)
}

func (mod *Module) addToSet(setRow *dbSet, dataID data.ID) error {
	dataRow, err := mod.dbDataFindOrCreateByDataID(dataID)
	if err != nil {
		return err
	}

	return mod.dbSetAddDataID([]uint{setRow.ID}, dataRow.ID)
}

func (mod *Module) RemoveFromSet(name string, dataID data.ID) error {
	setRow, err := mod.dbFindSetByName(name)
	if err != nil {
		return err
	}
	if sets.Type(setRow.Type) != sets.TypeSet {
		return sets.ErrInvalidSetType
	}

	return mod.removeFromSet(setRow, dataID)
}

func (mod *Module) removeFromSet(setRow *dbSet, dataID data.ID) error {
	dataRow, err := mod.dbDataFindOrCreateByDataID(dataID)
	if err != nil {
		return err
	}

	return mod.dbSetRemoveDataID([]uint{setRow.ID}, dataRow.ID)
}

func (mod *Module) SetInfo(name string) (*sets.Info, error) {
	setRow, err := mod.dbFindSetByName(name)
	if err != nil {
		return nil, err
	}

	var info = &sets.Info{
		Name:        setRow.Name,
		Type:        sets.Type(setRow.Type),
		Size:        0,
		Visible:     setRow.Visible,
		Description: setRow.Description,
		CreatedAt:   setRow.CreatedAt,
	}

	var count int64

	var tx = mod.db.
		Model(&dbMember{}).
		Where("set_id = ? and added = true", setRow.ID).
		Count(&count)

	if tx.Error != nil {
		return nil, tx.Error
	}

	info.Size = int(count)

	return info, nil
}

func (mod *Module) Where(dataID data.ID) ([]string, error) {
	dataRow, err := mod.dbDataFindOrCreateByDataID(dataID)
	if err != nil {
		return nil, err
	}

	rows, err := mod.dbMemberFindByDataID(dataRow.ID)
	if err != nil {
		return nil, err
	}

	var list []string

	for _, row := range rows {
		if row.Added {
			list = append(list, row.Set.Name)
		}
	}

	return list, nil
}

func (mod *Module) AllSets() ([]sets.Info, error) {
	var list []sets.Info

	rows, err := mod.dbSetFindAll()
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		var count int64

		var tx = mod.db.
			Model(&dbMember{}).
			Where("set_id = ? and added = true", row.ID).
			Count(&count)

		if tx.Error != nil {
			mod.log.Errorv(2, "error getting entry count: %v", tx.Error)
			count = -1
		}

		list = append(list, sets.Info{
			Name:        row.Name,
			Type:        sets.Type(row.Type),
			Size:        int(count),
			Visible:     row.Visible,
			Description: row.Description,
			CreatedAt:   row.CreatedAt,
		})
	}

	return list, err
}

func (mod *Module) UpdatedBetween(name string, since time.Time, until time.Time) ([]sets.Member, error) {
	setRow, err := mod.dbFindSetByName(name)
	if err != nil {
		return nil, err
	}

	rows, err := mod.dbMemberFindUpdatedBetween(setRow.ID, since, until)
	if err != nil {
		return nil, err
	}

	var updates []sets.Member

	for _, row := range rows {
		updates = append(updates, sets.Member{
			DataID:    row.Data.DataID,
			Added:     row.Added,
			UpdatedAt: row.UpdatedAt,
		})
	}

	return updates, nil
}

func (mod *Module) Scan(name string, opts *sets.ScanOpts) ([]*sets.Member, error) {
	setRow, err := mod.dbFindSetByName(name)
	if err != nil {
		return nil, err
	}

	if opts == nil {
		opts = &sets.ScanOpts{}
	}

	var rows []dbMember
	var q = mod.db.
		Where("set_id = ?", setRow.ID).
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

	var entries []*sets.Member

	for _, row := range rows {
		entries = append(entries, &sets.Member{
			DataID:    row.Data.DataID,
			Added:     row.Added,
			UpdatedAt: row.UpdatedAt,
		})
	}

	return entries, nil
}

func (mod *Module) AddToUnion(superset string, subset string) error {
	unionRow, err := mod.dbFindSetByName(superset)
	if err != nil {
		return err
	}

	setRow, err := mod.dbFindSetByName(subset)
	if err != nil {
		return err
	}

	err = mod.dbUnionCreate(unionRow.ID, setRow.ID)
	if err != nil {
		return err
	}

	entries, err := mod.dbMemberFindUpdatedBetween(setRow.ID, time.Time{}, time.Time{})
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.Added {
			continue
		}

		err = mod.addToSet(unionRow, entry.Data.DataID)
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

func (mod *Module) RemoveFromUnion(superset string, subset string) error {
	unionRow, err := mod.dbFindSetByName(superset)
	if err != nil {
		return err
	}

	setRow, err := mod.dbFindSetByName(subset)
	if err != nil {
		return err
	}

	err = mod.dbUnionDelete(unionRow.ID, setRow.ID)
	if err != nil {
		return err
	}

	var dataIDs []uint

	err = mod.db.
		Model(&dbMember{}).
		Where("set_id = ? and added = true", setRow.ID).
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
			mod.dbSetRemoveDataID([]uint{unionRow.ID}, dataID)
		}
	}

	return nil
}

func (mod *Module) Contains(name string, dataID data.ID) (bool, error) {
	setRow, err := mod.dbFindSetByName(name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, sets.ErrSetNotFound
		}
		return false, fmt.Errorf("cannot read set %s: %w", name, err)
	}

	dataRow, err := mod.dbDataFindOrCreateByDataID(dataID)
	if err != nil {
		return false, err
	}

	entry, err := mod.dbMemberFind(setRow.ID, dataRow.ID)

	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return false, nil
	case err != nil:
		return false, err
	default:
		return entry.Added, nil
	}
}

// Member - find a set member
// Errors: sets.ErrSetNotFound sets.ErrMemberNotFound sets.ErrDatabaseError
func (mod *Module) Member(name string, dataID data.ID) (*sets.Member, error) {
	setRow, err := mod.dbFindSetByName(name)
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil, sets.ErrSetNotFound
	case err != nil:
		return nil, sets.DatabaseError(err)
	}

	dataRow, err := mod.dbDataFindOrCreateByDataID(dataID)
	if err != nil {
		return nil, sets.DatabaseError(err)
	}

	row, err := mod.dbMemberFind(setRow.ID, dataRow.ID)
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil, sets.ErrMemberNotFound
	case err != nil:
		return nil, sets.DatabaseError(err)
	}

	return &sets.Member{
		DataID:    dataID,
		Added:     row.Added,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

func (mod *Module) SetVisible(name string, visible bool) error {
	return mod.dbSetUpdateVisible(name, visible)
}

func (mod *Module) SetDescription(name string, desc string) error {
	return mod.dbSetSetDescription(name, desc)
}
