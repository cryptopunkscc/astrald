package objects

import (
	"cmp"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
	"slices"
)

var defaultReadOpts = &objects.OpenOpts{Virtual: true}

type Opener struct {
	objects.Opener
	Name     string
	Priority int
}

func (mod *Module) Open(objectID object.ID, opts *objects.OpenOpts) (objects.Reader, error) {
	if opts == nil {
		opts = defaultReadOpts
	}

	openers := mod.openers.Values()

	slices.SortFunc(openers, func(a, b *Opener) int {
		return cmp.Compare(a.Priority, b.Priority) * -1 // from high to low
	})

	for _, opener := range openers {
		r, err := opener.Open(objectID, opts)
		if err == nil {
			return r, nil
		}
	}

	return nil, objects.ErrNotFound
}

func (mod *Module) AddOpener(name string, opener objects.Opener, priority int) error {
	_, ok := mod.openers.Set(name, &Opener{
		Opener:   opener,
		Name:     name,
		Priority: priority,
	})
	if !ok {
		return objects.ErrAlreadyExists
	}
	return nil
}
