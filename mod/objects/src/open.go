package objects

import (
	"cmp"
	"context"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
	"slices"
	"time"
)

const openTimeout = 30 * time.Second

type Opener struct {
	objects.Opener
	Priority int
}

type OpenerSet []*Opener

func (mod *Module) Open(ctx context.Context, objectID object.ID, opts *objects.OpenOpts) (objects.Reader, error) {
	if opts == nil {
		opts = &objects.OpenOpts{}
		*opts = *objects.DefaultOpenOpts
	}

	zone := opts.Zone
	defer func() {
		opts.Zone = zone
	}()

	openers := OpenerSet(mod.openers.Clone())
	slices.SortFunc(openers, func(a, b *Opener) int {
		return cmp.Compare(a.Priority, b.Priority) * -1 // from high to low
	})

	ctx, cancel := context.WithTimeout(ctx, openTimeout)
	defer cancel()

	// limit first attempt to local zone
	if zone.Is(objects.ZoneLocal) {
		opts.Zone = zone & objects.ZoneLocal
		r, err := openers.OpenFirst(ctx, objectID, opts)
		if err == nil {
			return r, nil
		}
	}

	// then include the virtual zone
	if zone.Is(objects.ZoneVirtual) {
		opts.Zone = zone & (objects.ZoneLocal | objects.ZoneVirtual)
		r, err := openers.OpenFirst(ctx, objectID, opts)
		if err == nil {
			return r, nil
		}
	}

	// then include the network
	if zone.Is(objects.ZoneNetwork) {
		opts.Zone = zone & (objects.ZoneLocal | objects.ZoneVirtual | objects.ZoneNetwork)
		r, err := openers.OpenFirst(ctx, objectID, opts)
		if err == nil {
			return r, nil
		}
	}

	return nil, objects.ErrNotFound
}

func (mod *Module) AddOpener(opener objects.Opener, priority int) error {
	return mod.openers.Add(&Opener{
		Opener:   opener,
		Priority: priority,
	})
}

func (set OpenerSet) OpenFirst(ctx context.Context, objectID object.ID, opts *objects.OpenOpts) (objects.Reader, error) {
	for _, opener := range set {
		r, err := opener.Open(ctx, objectID, opts)
		if err == nil {
			return r, nil
		}
	}

	return nil, objects.ErrNotFound
}
