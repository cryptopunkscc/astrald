package sets

import (
	"context"
	"fmt"
	"github.com/cryptopunkscc/astrald/mod/sets"
	"gorm.io/gorm"
)

var _ sets.Union = &UnionSet{}

type UnionSet struct {
	*Module
	row  *dbSet
	edit sets.Editor
}

func (mod *Module) unionOpener(name string) (*UnionSet, error) {
	var err error
	var set = &UnionSet{Module: mod}

	set.row, err = mod.dbFindSetByName(name)
	if err != nil {
		return nil, err
	}

	set.edit, err = mod.Edit(name)
	if err != nil {
		return nil, err
	}

	return set, nil
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
		super, err := sets.Open[*UnionSet](mod, row.Name)
		if err != nil {
			continue
		}
		super.Sync()
	}

	return nil
}

func (set *UnionSet) Add(names ...string) error {
	defer set.Sync()

	for _, name := range names {
		subsetRow, err := set.dbFindSetByName(name)
		if err != nil {
			return err
		}

		err = set.db.Create(&dbSetInclusion{
			SupersetID: set.row.ID,
			SubsetID:   subsetRow.ID,
		}).Error
		if err != nil {
			return err
		}
	}

	set.events.Emit(eventSetUpdated{row: set.row})
	set.events.Emit(sets.EventSetUpdated{Name: set.row.Name})

	return nil
}

func (set *UnionSet) Remove(names ...string) error {
	defer set.Sync()

	for _, name := range names {
		subsetRow, err := set.dbFindSetByName(name)
		if err != nil {
			return err
		}

		err = set.db.Delete(&dbSetInclusion{
			SupersetID: set.row.ID,
			SubsetID:   subsetRow.ID,
		}).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func (set *UnionSet) Subsets() ([]string, error) {
	var list []string

	var err = set.db.
		Model(&dbSet{}).
		Select("name").
		Where(
			"id IN (?)",
			set.db.
				Model(&dbSetInclusion{}).
				Where("superset_id = ?", set.row.ID).
				Select("subset_id"),
		).
		Find(&list).Error

	return list, err
}

func (set *UnionSet) Sync() error {
	var err error

	err = set.deleteExcess()
	if err != nil {
		return err
	}

	return set.addMissing()
}

func (set *UnionSet) Scan(opts *sets.ScanOpts) ([]*sets.Member, error) {
	return set.edit.Scan(opts)
}

func (set *UnionSet) Info() (*sets.Info, error) {
	var info = &sets.Info{
		Name:        set.row.Name,
		Type:        sets.Type(set.row.Type),
		Size:        -1,
		Visible:     set.row.Visible,
		Description: set.row.Description,
		CreatedAt:   set.row.CreatedAt,
		TrimmedAt:   set.row.TrimmedAt,
	}

	var count int64

	var tx = set.db.
		Model(&dbMember{}).
		Where("set_id = ? and removed = false", set.row.ID).
		Count(&count)

	if tx.Error != nil {
		return nil, tx.Error
	}

	info.Size = int(count)

	return info, nil
}

func (set *UnionSet) deleteExcess() error {
	var err error
	var ids []uint

	// find excess members
	err = set.dbExcessMembers().Find(&ids).Error
	if err != nil {
		return fmt.Errorf("select error: %w", err)
	}

	return set.edit.RemoveByID(ids...)
}

func (set *UnionSet) addMissing() error {
	var err error
	var missingIDs []uint

	err = set.dbMissingMembers().Find(&missingIDs).Error
	if err != nil {
		return err
	}

	return set.edit.AddByID(missingIDs...)
}

func (set *UnionSet) dbSubsets() *gorm.DB {
	return set.db.
		Model(&dbSetInclusion{}).
		Select("subset_id").
		Where("superset_id = ?", set.row.ID)
}

func (set *UnionSet) dbMembers() *gorm.DB {
	return set.db.
		Model(&dbMember{}).
		Select("data_id").
		Where("removed = false AND set_id = ?", set.row.ID)
}

func (set *UnionSet) dbSubsetMembers() *gorm.DB {
	return set.db.
		Model(&dbMember{}).
		Select("data_id").
		Distinct("data_id").
		Where("removed = false AND set_id IN (?)", set.dbSubsets())
}

func (set *UnionSet) dbExcessMembers() *gorm.DB {
	return set.dbMembers().
		Where("removed = false AND data_id NOT in (?)", set.dbSubsetMembers())
}

func (set *UnionSet) dbMissingMembers() *gorm.DB {
	return set.dbSubsetMembers().
		Where("data_id NOT IN (?)", set.dbMembers())
}
