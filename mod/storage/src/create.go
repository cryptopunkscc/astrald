package storage

import (
	"cmp"
	"errors"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"slices"
)

type Creator struct {
	storage.Creator
	Name     string
	Priority int
}

func (mod *Module) Create(opts *storage.CreateOpts) (storage.Writer, error) {
	if opts == nil {
		opts = &storage.CreateOpts{}
	}

	if opts.Alloc < 0 {
		return nil, errors.New("alloc cannot be less than 0")
	}

	creators := mod.creators.Values()

	slices.SortFunc(creators, func(a, b *Creator) int {
		return cmp.Compare(a.Priority, b.Priority) * -1 // from high to low
	})

	for _, creator := range creators {
		w, err := creator.Create(opts)
		if err == nil {
			return NewDataWriterWrapper(mod, w), err
		}
	}

	return nil, storage.ErrStorageUnavailable
}
func (mod *Module) AddCreator(name string, creator storage.Creator, priority int) error {
	_, ok := mod.creators.Set(name, &Creator{
		Creator:  creator,
		Name:     name,
		Priority: priority,
	})

	if ok {
		mod.events.Emit(storage.EventStoreAdded{
			Name:    name,
			Creator: creator,
		})
		return nil
	}
	return storage.ErrAlreadyExists
}
func (mod *Module) RemoveCreator(name string) error {
	if creator, ok := mod.creators.Delete(name); ok {
		mod.events.Emit(storage.EventStoreRemoved{
			Name:    name,
			Creator: creator,
		})
	}
	return storage.ErrNotFound
}
