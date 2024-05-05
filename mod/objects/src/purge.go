package objects

import (
	"errors"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
)

func (mod *Module) Purge(objectID object.ID, opts *objects.PurgeOpts) (int, error) {
	var total int
	var errs []error

	if opts == nil {
		opts = &objects.PurgeOpts{}
	}

	for _, purger := range mod.purgers.Clone() {
		n, err := purger.Purge(objectID, opts)
		if err != nil {
			errs = append(errs, err)
		}
		total += n
	}

	if total > 0 {
		mod.events.Emit(objects.EventObjectPurged{ObjectID: objectID})
	}

	return total, errors.Join(errs...)
}

func (mod *Module) AddPurger(purger objects.Purger) error {
	return mod.purgers.Add(purger)
}
