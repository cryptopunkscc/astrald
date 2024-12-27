package objects

import (
	"cmp"
	"errors"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"slices"
)

type Creator struct {
	objects.Creator
	Priority int
}

func (mod *Module) Create(opts *objects.CreateOpts) (objects.Writer, error) {
	if opts == nil {
		opts = &objects.CreateOpts{}
	}

	if opts.Alloc < 0 {
		return nil, errors.New("alloc cannot be less than 0")
	}

	creators := mod.creators.Clone()

	slices.SortFunc(creators, func(a, b *Creator) int {
		return cmp.Compare(a.Priority, b.Priority) * -1 // from high to low
	})

	for _, creator := range creators {
		w, err := creator.CreateObject(opts)
		if err == nil {
			return NewWriterWrapper(mod, w), err
		}
	}

	return nil, objects.ErrStorageUnavailable
}

func (mod *Module) AddCreator(creator objects.Creator, priority int) error {
	return mod.creators.Add(&Creator{
		Creator:  creator,
		Priority: priority,
	})
}
