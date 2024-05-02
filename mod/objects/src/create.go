package objects

import (
	"cmp"
	"errors"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"slices"
)

type Creator struct {
	objects.Creator
	Name     string
	Priority int
}

func (mod *Module) Create(opts *objects.CreateOpts) (objects.Writer, error) {
	if opts == nil {
		opts = &objects.CreateOpts{}
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
			return NewWriterWrapper(mod, w), err
		}
	}

	return nil, objects.ErrStorageUnavailable
}

func (mod *Module) AddCreator(name string, creator objects.Creator, priority int) error {
	_, ok := mod.creators.Set(name, &Creator{
		Creator:  creator,
		Name:     name,
		Priority: priority,
	})
	if !ok {
		return objects.ErrAlreadyExists
	}
	return nil
}
