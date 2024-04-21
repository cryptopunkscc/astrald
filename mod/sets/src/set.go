package sets

import (
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/sets"
	"github.com/cryptopunkscc/astrald/sig"
	"time"
)

var _ sets.Set = &Set{}

type Set struct {
	*Module
	row *dbSet
}

func (set *Set) Name() string {
	return set.row.Name
}

func (set *Set) Scan(opts *sets.ScanOpts) ([]*sets.Member, error) {
	if opts == nil {
		opts = &sets.ScanOpts{}
	}

	var rows []dbMember
	var q = set.db.
		Where("set_id = ?", set.row.ID).
		Order("updated_at").
		Preload("Data")

	if !opts.UpdatedAfter.IsZero() {
		q = q.Where("updated_at > ?", opts.UpdatedAfter)
	}
	if !opts.UpdatedBefore.IsZero() {
		q = q.Where("updated_at < ?", opts.UpdatedBefore)
	}
	if !opts.IncludeRemoved {
		q = q.Where("removed = false")
	}
	if !opts.DataID.IsZero() {
		row, err := set.dbDataFindByDataID(opts.DataID)
		if err != nil {
			return nil, err
		}
		q = q.Where("data_id = ?", row.ID)
	}

	err := q.Find(&rows).Error
	if err != nil {
		return nil, err
	}

	var entries []*sets.Member

	for _, row := range rows {
		entries = append(entries, &sets.Member{
			DataID:    row.Data.DataID,
			Removed:   row.Removed,
			UpdatedAt: row.UpdatedAt,
		})
	}

	return entries, nil
}

func (set *Set) Add(dataIDs ...data.ID) error {
	var ids []uint

	for _, dataID := range dataIDs {
		row, err := set.dbDataFindOrCreateByDataID(dataID)
		if err != nil {
			return err
		}
		ids = append(ids, row.ID)
	}
	return set.AddByID(ids...)
}

func (set *Set) Remove(dataIDs ...data.ID) error {
	var ids []uint

	for _, dataID := range dataIDs {
		row, err := set.dbDataFindOrCreateByDataID(dataID)
		if err != nil {
			return err
		}
		ids = append(ids, row.ID)
	}
	return set.RemoveByID(ids...)
}

func (set *Set) RemoveByID(ids ...uint) error {
	if len(ids) == 0 {
		return nil
	}

	var err error
	var clean []uint

	// reduce the id set to members actually in the set
	err = set.db.
		Model(&dbMember{}).
		Select("data_id").
		Where("removed = false AND set_id = ? AND data_id IN (?)", set.row.ID, ids).
		Find(&clean).
		Error

	// update members as removed from the set
	err = set.db.
		Model(&dbMember{}).
		Where("removed = false AND set_id = ? AND data_id IN (?)", set.row.ID, clean).
		Update("removed", true).Error
	if err != nil {
		return fmt.Errorf("update error: %w", err)
	}

	// fetch details about the removed rows
	var rows []dbMember
	err = set.db.
		Preload("Data").
		Where("set_id = ? AND data_id IN (?)", set.row.ID, clean).
		Find(&rows).Error
	if err != nil {
		return err
	}

	// emit an event for every removed member
	for _, row := range rows {
		set.events.Emit(sets.EventMemberUpdate{
			Set:       set.row.Name,
			DataID:    row.Data.DataID,
			Removed:   row.Removed,
			UpdatedAt: row.UpdatedAt,
		})
	}

	set.events.Emit(sets.EventSetUpdated{Name: set.row.Name})

	return nil
}

func (set *Set) AddByID(ids ...uint) error {
	var err error
	var duplicates []uint
	var removedIDs []uint

	// filter out elements that are already added
	err = set.db.
		Model(&dbMember{}).
		Select("data_id").
		Where("removed = false AND set_id = ? AND data_id IN (?)", set.row.ID, ids).
		Find(&duplicates).Error
	if err != nil {
		return err
	}

	ids = subtract(ids, duplicates)

	// fetch existing rows marked as removed
	err = set.db.
		Model(&dbMember{}).
		Select("data_id").
		Where("removed = true AND set_id = ? AND data_id IN (?)", set.row.ID, ids).
		Find(&removedIDs).Error
	if err != nil {
		return err
	}

	var insertIDs = subtract(ids, removedIDs)

	err = set.unremove(removedIDs)
	if err != nil {
		return err
	}

	err = set.insert(insertIDs)
	if err != nil {
		return err
	}

	set.events.Emit(sets.EventSetUpdated{Name: set.row.Name})

	return nil
}

func (set *Set) Stat() (*sets.Stat, error) {
	var err error
	var info = &sets.Stat{
		Name:      set.row.Name,
		Size:      -1,
		CreatedAt: set.row.CreatedAt,
		TrimmedAt: set.row.TrimmedAt,
	}

	var rows []dbMember

	err = set.db.
		Model(&dbMember{}).
		Preload("Data").
		Where("set_id = ? and removed = false", set.row.ID).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		info.DataSize += row.Data.DataID.Size
	}

	info.Size = len(rows)

	return info, nil
}

func (set *Set) Trim(t time.Time) error {
	if t.After(time.Now()) {
		return errors.New("invalid time")
	}

	if !t.After(set.row.TrimmedAt) {
		return errors.New("already trimmed with later date")
	}

	set.row.TrimmedAt = t
	err := set.db.Save(set.row).Error
	if err != nil {
		return err
	}

	err = set.db.
		Where("removed = true AND updated_at < ?", t).
		Delete(&dbMember{}).Error

	return err
}

func (set *Set) TrimmedAt() time.Time {
	return set.row.TrimmedAt
}

func (set *Set) Clear() error {
	var err error
	var ids []uint

	err = set.db.
		Model(&dbMember{}).
		Select("data_id").
		Where("set_id = ?", set.row.ID).
		Find(&ids).Error
	if err != nil {
		return err
	}

	err = set.RemoveByID(ids...)
	if err != nil {
		return err
	}

	set.row.TrimmedAt = time.Now()
	err = set.db.Save(set.row).Error

	set.events.Emit(sets.EventSetUpdated{Name: set.row.Name})

	return err
}

func (set *Set) Delete() error {
	var err error
	var rows []dbMember

	var lastState dbSet
	err = set.db.
		Preload("InclusionsAsSuper").
		Preload("InclusionsAsSub").
		Where("name = ?", set.row.Name).
		First(&lastState).
		Error

	// fetch members
	err = set.db.
		Preload("Data").
		Where("removed = false AND set_id = ?", set.row.ID).
		Find(&rows).Error
	if err != nil {
		return err
	}

	err = set.db.
		Where("set_id = ?", set.row.ID).
		Delete(&dbMember{}).Error
	if err != nil {
		return err
	}

	// emit an event for every removed member
	for _, row := range rows {
		set.events.Emit(sets.EventMemberUpdate{
			Set:       set.row.Name,
			DataID:    row.Data.DataID,
			Removed:   true,
			UpdatedAt: row.UpdatedAt,
		})
	}

	err = set.db.Delete(set.row).Error
	if err != nil {
		return err
	}

	set.events.Emit(sets.EventSetDeleted{Name: set.row.Name})

	return nil
}

func (set *Set) unremove(ids []uint) error {
	if len(ids) == 0 {
		return nil
	}

	var err error

	// update rows
	err = set.db.
		Model(&dbMember{}).
		Where("removed = true AND set_id = ? AND data_id IN (?)", set.row.ID, ids).
		Update("removed", false).Error
	if err != nil {
		return fmt.Errorf("update error: %w", err)
	}

	var rows []dbMember

	// reload rows to include DataID
	err = set.db.
		Preload("Data").
		Where("set_id = ? AND data_id IN (?)", set.row.ID, ids).
		Find(&rows).Error
	if err != nil {
		return err
	}

	// emit an event for every updated member
	for _, row := range rows {
		set.events.Emit(sets.EventMemberUpdate{
			Set:       set.row.Name,
			DataID:    row.Data.DataID,
			Removed:   row.Removed,
			UpdatedAt: row.UpdatedAt,
		})
	}

	return nil
}

func (set *Set) insert(ids []uint) error {
	if len(ids) == 0 {
		return nil
	}

	var err error

	// create a slice of rows
	rows, _ := sig.MapSlice(ids, func(id uint) (*dbMember, error) {
		return &dbMember{
			DataID: id,
			SetID:  set.row.ID,
		}, nil
	})

	err = set.db.Create(&rows).Error
	if err != nil {
		return err
	}

	// reload rows to include Data
	err = set.db.
		Preload("Data").
		Where("set_id = ? AND data_id IN (?)", set.row.ID, ids).
		Find(&rows).Error
	if err != nil {
		return err
	}

	// emit an event for every updated  member
	for _, row := range rows {
		set.events.Emit(sets.EventMemberUpdate{
			Set:       set.row.Name,
			DataID:    row.Data.DataID,
			Removed:   row.Removed,
			UpdatedAt: row.UpdatedAt,
		})
	}
	return nil
}

func subtract(set, subset []uint) []uint {
	var res []uint
	var i, j int

	for i < len(set) && j < len(subset) {
		if set[i] < subset[j] {
			res = append(res, set[i])
			i++
		} else if set[i] > subset[j] {
			j++
		} else {
			// If elements are equal, move to the next element in arr1
			i++
			j++
		}
	}

	// Append the remaining elements from set
	res = append(res, set[i:]...)

	return res
}
