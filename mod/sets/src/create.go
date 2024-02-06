package sets

import (
	"errors"
	"github.com/cryptopunkscc/astrald/mod/sets"
)

func (mod *Module) Create(name string) (sets.Set, error) {
	return mod.CreateManaged(name, sets.TypeBasic)
}

func (mod *Module) CreateManaged(name string, typ sets.Type) (sets.Set, error) {
	wrapper, found := mod.wrappers.Get(string(typ))
	if !found {
		return nil, errors.New("unsupported set type")
	}

	var row = dbSet{
		Name: name,
		Type: string(typ),
	}

	err := mod.db.Create(&row).Error
	if err != nil {
		return nil, err
	}

	var set = &Set{
		Module: mod,
		row:    &row,
	}

	wrapped, err := wrapper(set)
	if err != nil {
		return nil, err
	}

	mod.events.Emit(sets.EventSetCreated{Set: set})

	return wrapped, nil
}

func (mod *Module) CreateUnion(name string) (sets.Union, error) {
	set, err := mod.CreateManaged(name, sets.TypeUnion)
	if err != nil {
		return nil, err
	}

	union, ok := set.(sets.Union)
	if !ok {
		return nil, errors.New("set is not a union")
	}

	return union, nil
}

func (mod *Module) Open(name string, create bool) (sets.Set, error) {
	var row dbSet
	var err = mod.db.Where("name = ?", name).First(&row).Error
	if err != nil {
		if create {
			return mod.Create(name)
		}
		return nil, err
	}

	var set = &Set{
		Module: mod,
		row:    &row,
	}

	wrapper, ok := mod.wrappers.Get(row.Type)
	if !ok {
		mod.log.Errorv(2, "set %v has an unknown set type: %v",
			row.Name,
			row.Type,
		)
		return nil, errors.New("invalid set type")
	}

	wrapped, err := wrapper(set)
	if err != nil {
		return nil, err
	}

	return wrapped, nil
}

func (mod *Module) OpenUnion(name string, create bool) (sets.Union, error) {
	set, err := mod.Open(name, false)
	if err != nil {
		if create {
			return mod.CreateUnion(name)
		}
		return nil, err
	}

	union, ok := set.(sets.Union)
	if !ok {
		return nil, errors.New("set is not a union")
	}

	return union, nil
}

func (mod *Module) OpenOrCreateUnion(name string) (sets.Union, error) {
	if s, err := mod.OpenUnion(name, false); err == nil {
		return s, nil
	}
	return mod.CreateUnion(name)
}
