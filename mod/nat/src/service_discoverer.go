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

			change := services.ServiceChange{
				Type:    services.ServiceChangeTypeSnapshot,
				Enabled: astral.Bool(mod.serviceEnabled),
				Service: svc,
			}
			select {
			case out <- change:
			case <-ctx.Done():
				return
			}

			// Signal end of snapshot phase for this discoverer.
			select {
			case out <- services.ServiceChange{Type: services.ServiceChangeTypeFlush}:
			case <-ctx.Done():
				return
			}
		}

		if !opts.Follow {
			return
		}

		sub := mod.serviceChangeFeed.Subscribe(ctx)
		for {
			select {
			case <-ctx.Done():
				return
			case ev, ok := <-sub:
				if !ok {
					return
				}

				// Normalize to update if the producer didn't set a type.
				if ev.Type == "" {
					ev.Type = services.ServiceChangeTypeUpdate
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
