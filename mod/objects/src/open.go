package objects

import (
	"cmp"
	"github.com/cryptopunkscc/astrald/astral"
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

func (mod *Module) Open(ctx *astral.Context, objectID object.ID, opts *objects.OpenOpts) (objects.Reader, error) {
	if opts == nil {
		opts = objects.DefaultOpenOpts()
	}

	// authorize if necessary
	if !ctx.Identity().IsEqual(mod.node.Identity()) {
		if !mod.Auth.Authorize(ctx.Identity(), objects.ActionRead, &objectID) {
			return nil, objects.ErrAccessDenied
		}
	}

	openers := OpenerSet(mod.openers.Clone())
	slices.SortFunc(openers, func(a, b *Opener) int {
		return cmp.Compare(a.Priority, b.Priority) * -1 // from high to low
	})

	ctx, cancel := ctx.WithTimeout(openTimeout)
	defer cancel()

	// limit first attempt to local zone
	if ctx.Zone().Is(astral.ZoneDevice) {
		r, err := openers.OpenFirst(ctx.LimitZones(astral.ZoneDevice), objectID, opts)
		if err == nil {
			return r, nil
		}
	}

	// then include the virtual zone
	if ctx.Zone().Is(astral.ZoneVirtual) {
		r, err := openers.OpenFirst(ctx.LimitZones(astral.ZoneDevice|astral.ZoneVirtual), objectID, opts)
		if err == nil {
			return r, nil
		}
	}

	// then try the network
	if ctx.Zone().Is(astral.ZoneNetwork) {
		r, err := mod.openNetwork(ctx, objectID, opts)
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

func (mod *Module) openNetwork(ctx *astral.Context, objectID object.ID, opts *objects.OpenOpts) (objects.Reader, error) {
	providers := mod.Find(ctx, objectID, &astral.Scope{Zone: ctx.Zone()})

	if opts.QueryFilter != nil {
		providers = slices.DeleteFunc(providers, func(identity *astral.Identity) bool {
			return !opts.QueryFilter(identity)
		})
	}

	var conns = make(chan objects.Reader, 1)
	var wg sync.WaitGroup

	ctx, done := ctx.WithCancel()
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

func (set OpenerSet) OpenFirst(ctx *astral.Context, objectID object.ID, opts *objects.OpenOpts) (objects.Reader, error) {
	for _, opener := range set {
		r, err := opener.OpenObject(ctx, objectID, opts)
		if err == nil {
			return r, nil
		}
	}

	return nil, objects.ErrNotFound
}
