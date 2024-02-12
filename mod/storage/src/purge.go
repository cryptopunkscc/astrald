package storage

import (
	"errors"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/storage"
)

func (mod *Module) Purge(dataID data.ID, opts *storage.PurgeOpts) (int, error) {
	var total int
	var errs []error

	if opts == nil {
		opts = &storage.PurgeOpts{}
	}

	for _, purger := range mod.purgers.Clone() {
		n, err := purger.Purge(dataID, opts)
		if err != nil {
			errs = append(errs, err)
		}
		total += n
	}

	if total > 0 {
		mod.events.Emit(storage.EventDataPurged{DataID: dataID})
	}

	return total, errors.Join(errs...)
}

func (mod *Module) AddPurger(name string, purger storage.Purger) error {
	_, ok := mod.purgers.Set(name, purger)
	if ok {
		mod.events.Emit(storage.EventPurgerAdded{
			Name:   name,
			Purger: purger,
		})
		return nil
	}
	return storage.ErrAlreadyExists
}

func (mod *Module) RemovePurger(name string) error {
	if purger, ok := mod.purgers.Delete(name); ok {

		mod.events.Emit(storage.EventPurgerRemoved{
			Name:   name,
			Purger: purger,
		})
	}
	return storage.ErrNotFound
}
