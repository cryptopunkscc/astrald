package nat

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nat"
	"github.com/cryptopunkscc/astrald/mod/services"
)

func (mod *Module) DiscoverService(
	ctx *astral.Context,
	caller *astral.Identity,
	opts services.DiscoverOptions,
) (<-chan services.ServiceChange, error) {
	// Snapshot = false, Follow = false: immediately close and emit nothing.
	if !opts.Snapshot && !opts.Follow {
		out := make(chan services.ServiceChange)
		close(out)
		return out, nil
	}

	out := make(chan services.ServiceChange, 16)

	go func() {
		defer close(out)

		if opts.Snapshot {
			svc := services.Service{
				Name:        nat.ModuleName,
				Identity:    mod.node.Identity(),
				Composition: astral.NewBundle(),
			}

			change := services.ServiceChange{Enabled: astral.Bool(mod.serviceEnabled), Service: svc}
			select {
			case out <- change:
			case <-ctx.Done():
				return
			}
		}

		// Snapshot-only behavior.
		if !opts.Follow {
			return
		}

		// Follow future updates exclusively from the ServiceFeed.
		sub := mod.serviceFeed.Subscribe(ctx)
		for {
			select {
			case <-ctx.Done():
				return
			case ev, ok := <-sub:
				if !ok {
					return
				}
				select {
				case out <- ev:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return out, nil
}
