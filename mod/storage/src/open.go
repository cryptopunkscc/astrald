package storage

import (
	"cmp"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"slices"
)

var defaultReadOpts = &storage.OpenOpts{Virtual: true}

type Opener struct {
	storage.Opener
	Name     string
	Priority int
}

func (mod *Module) Open(dataID data.ID, opts *storage.OpenOpts) (storage.Reader, error) {
	if opts == nil {
		opts = defaultReadOpts
	}

	openers := mod.openers.Values()

	slices.SortFunc(openers, func(a, b *Opener) int {
		return cmp.Compare(a.Priority, b.Priority) * -1 // from high to low
	})

	for _, opener := range openers {
		r, err := opener.Open(dataID, opts)
		if err == nil {
			return r, nil
		}
	}

	return nil, storage.ErrNotFound
}

func (mod *Module) AddOpener(name string, opener storage.Opener, priority int) error {
	_, ok := mod.openers.Set(name, &Opener{
		Opener:   opener,
		Name:     name,
		Priority: priority,
	})
	if ok {
		mod.events.Emit(storage.EventOpenerAdded{
			Name:   name,
			Opener: opener,
		})
		return nil
	}
	return storage.ErrAlreadyExists
}

func (mod *Module) RemoveOpener(name string) error {
	if opener, ok := mod.openers.Delete(name); ok {
		mod.events.Emit(storage.EventReaderRemoved{
			Name:   name,
			Opener: opener,
		})
	}
	return storage.ErrNotFound
}
