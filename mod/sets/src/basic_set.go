package sets

import (
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/sets"
)

var _ sets.Basic = &BasicSet{}

type BasicSet struct {
	*Module
	row  *dbSet
	edit sets.Editor
}

func (mod *Module) basicOpener(name string) (*BasicSet, error) {
	var err error
	var set = &BasicSet{Module: mod}

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

func (set *BasicSet) Scan(opts *sets.ScanOpts) ([]*sets.Member, error) {
	return set.edit.Scan(opts)
}

func (set *BasicSet) Add(dataID ...data.ID) error {
	return set.edit.Add(dataID...)
}

func (set *BasicSet) Remove(dataID ...data.ID) error {
	return set.edit.Remove(dataID...)

}

func (set *BasicSet) Info() (*sets.Info, error) {
	var info = &sets.Info{
		Name:        set.row.Name,
		Type:        sets.Type(set.row.Type),
		Size:        -1,
		Visible:     set.row.Visible,
		Description: set.row.Description,
		CreatedAt:   set.row.CreatedAt,
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
