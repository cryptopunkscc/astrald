package objects

import (
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

func (mod *Module) Purge(objectID *astral.ObjectID, opts *objects.PurgeOpts) (int, error) {
	var total int
	var errs []error

	if opts == nil {
		opts = &objects.PurgeOpts{}
	}

	for _, purger := range mod.purgers.Clone() {
		n, err := purger.PurgeObject(objectID, opts)
		if err != nil {
			errs = append(errs, err)
		}
		total += n
	}

	return total, errors.Join(errs...)
}

func (mod *Module) AddPurger(purger objects.Purger) error {
	return mod.purgers.Add(purger)
}
