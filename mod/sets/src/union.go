package sets

import (
	"context"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/mod/sets"
	"gorm.io/gorm"
)

var _ sets.Union = &Union{}

type Union struct {
	sets.Set
	mod *Module
	row *dbSet
}

func (mod *Module) unionWrapper(set sets.Set) (sets.Set, error) {
	var err error

	var union = &Union{
		Set: set,
		mod: mod,
	}

	union.row, err = mod.dbFindSetByName(set.Name())
	if err != nil {
		return nil, err
	}

	return union, nil
}

func (mod *Module) watchUnionMembers(ctx context.Context) error {
	for event := range mod.events.Subscribe(ctx) {
		switch e := event.(type) {
		case eventSetDeleted:
			var err = mod.syncInclusions(e.row.InclusionsAsSub...)
			if err != nil {
				mod.log.Errorv(2, "syncInclusions: %v", err)
			}
		case eventSetUpdated:
			var err = mod.syncSupersetsOf(e.row.ID)
			if err != nil {
				mod.log.Errorv(2, "syncInclusions: %v", err)
			}
		}
	}

	return nil
}

func (mod *Module) syncSupersetsOf(subsetID uint) error {
	var rows []dbSetInclusion
	var err = mod.db.
		Where("subset_id = ?", subsetID).
		Find(&rows).Error
	if err != nil {
		return err
	}

	return mod.syncInclusions(rows...)
}

func (mod *Module) syncInclusions(inclusions ...dbSetInclusion) error {
	var ids []uint
	for _, i := range inclusions {
		ids = append(ids, i.SupersetID)
	}

	var rows []dbSet
	var err = mod.db.
		Where("id IN (?)", ids).
		Find(&rows).Error
	if err != nil {
		return err
	}

	for _, row := range rows {
		super, err := mod.OpenUnion(row.Name, false)
		if err != nil {
			continue
		}
		super.Sync()
	}

	return nil
}

func (set *Union) AddSubset(names ...string) error {
	defer set.Sync()
	var errs []error

	for _, name := range names {
		subsetRow, err := set.mod.dbFindSetByName(name)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		err = set.mod.db.Create(&dbSetInclusion{
			SupersetID: set.row.ID,
			SubsetID:   subsetRow.ID,
		}).Error
		if err != nil {
			errs = append(errs, err)
			continue
		}
	}

	set.mod.events.Emit(eventSetUpdated{row: set.row})
	set.mod.events.Emit(sets.EventSetUpdated{Name: set.row.Name})

	return errors.Join(errs...)
}

func (set *Union) RemoveSubset(names ...string) error {
	defer set.Sync()
	var errs []error

	for _, name := range names {
		subsetRow, err := set.mod.dbFindSetByName(name)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		err = set.mod.db.Delete(&dbSetInclusion{
			SupersetID: set.row.ID,
			SubsetID:   subsetRow.ID,
		}).Error
		if err != nil {
			errs = append(errs, err)
			continue
		}
	}

	return errors.Join(errs...)
}

func (set *Union) Subsets() ([]string, error) {
	var list []string

	var err = set.mod.db.
		Model(&dbSet{}).
		Select("name").
		Where(
			"id IN (?)",
			set.mod.db.
				Model(&dbSetInclusion{}).
				Where("superset_id = ?", set.row.ID).
				Select("subset_id"),
		).
		Find(&list).Error

	return list, err
}

func (set *Union) Sync() error {
	var err error

	err = set.deleteExcess()
	if err != nil {
		return err
	}

	return set.addMissing()
}

func (set *Union) deleteExcess() error {
	var err error
	var ids []uint

	// find excess members
	err = set.dbExcessMembers().Find(&ids).Error
	if err != nil {
		return fmt.Errorf("select error: %w", err)
	}

	return set.Set.RemoveByID(ids...)
}

func (set *Union) addMissing() error {
	var err error
	var missingIDs []uint

	err = set.dbMissingMembers().Find(&missingIDs).Error
	if err != nil {
		return err
	}

	return set.Set.AddByID(missingIDs...)
}

func (set *Union) dbSubsets() *gorm.DB {
	return set.mod.db.
		Model(&dbSetInclusion{}).
		Select("subset_id").
		Where("superset_id = ?", set.row.ID)
}

func (set *Union) dbMembers() *gorm.DB {
	return set.mod.db.
		Model(&dbMember{}).
		Select("data_id").
		Where("removed = false AND set_id = ?", set.row.ID)
}

func (set *Union) dbSubsetMembers() *gorm.DB {
	return set.mod.db.
		Model(&dbMember{}).
		Select("data_id").
		Distinct("data_id").
		Where("removed = false AND set_id IN (?)", set.dbSubsets())
}

func (set *Union) dbExcessMembers() *gorm.DB {
	return set.dbMembers().
		Where("removed = false AND data_id NOT in (?)", set.dbSubsetMembers())
}

func (set *Union) dbMissingMembers() *gorm.DB {
	return set.dbSubsetMembers().
		Where("data_id NOT IN (?)", set.dbMembers())
}
