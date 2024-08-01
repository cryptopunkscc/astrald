package objects

import (
	"cmp"
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
	"slices"
	"sync"
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
		opts = objects.DefaultOpenOpts()
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
	if zone.Is(astral.ZoneDevice) {
		opts.Zone = zone & astral.ZoneDevice
		r, err := openers.OpenFirst(ctx, objectID, opts)
		if err == nil {
			return r, nil
		}
	}

	// then include the virtual zone
	if zone.Is(astral.ZoneVirtual) {
		opts.Zone = zone & (astral.ZoneDevice | astral.ZoneVirtual)
		r, err := openers.OpenFirst(ctx, objectID, opts)
		if err == nil {
			return r, nil
		}
	}

	// then try the network
	if zone.Is(astral.ZoneNetwork) {
		r, err := mod.openNetwork(ctx, objectID, opts)
		if err == nil {
			return r, nil
		}
	}

	return nil, objects.ErrNotFound
}

func (mod *Module) OpenAs(ctx context.Context, consumer id.Identity, objectID object.ID, opts *objects.OpenOpts) (objects.Reader, error) {
	if !mod.Auth.Authorize(consumer, objects.ActionRead, &objectID) {
		return nil, objects.ErrAccessDenied
	}

	return mod.Open(ctx, objectID, opts)
}

func (mod *Module) AddOpener(opener objects.Opener, priority int) error {
	return mod.openers.Add(&Opener{
		Opener:   opener,
		Priority: priority,
	})
}

func (mod *Module) openNetwork(ctx context.Context, objectID object.ID, opts *objects.OpenOpts) (objects.Reader, error) {
	if !opts.Zone.Is(astral.ZoneNetwork) {
		return nil, astral.ErrZoneExcluded
	}

	providers := mod.Find(ctx, objectID, &astral.Scope{Zone: opts.Zone})

	if opts.QueryFilter != nil {
		providers = slices.DeleteFunc(providers, func(identity id.Identity) bool {
			return !opts.QueryFilter(identity)
		})
	}

	var conns = make(chan objects.Reader, 1)
	var wg sync.WaitGroup

	ctx, done := context.WithCancel(ctx)
	defer done()

	for _, providerID := range providers {
		providerID := providerID

		wg.Add(1)
		go func() {
			defer wg.Done()

			c := NewConsumer(mod, mod.node.Identity(), providerID)

			r, err := c.Open(ctx, objectID, opts)
			if err != nil {
				return
			}

			select {
			case conns <- r:
				done()
			default:
				r.Close()
			}
		}()
	}

	wg.Wait()
	close(conns)

	r, ok := <-conns
	if !ok {
		return nil, objects.ErrNotFound
	}

	return r, nil
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
