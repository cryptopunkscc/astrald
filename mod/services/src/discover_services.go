package services

import (
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/services"
)

// DiscoverServices runs all registered ServiceDiscoverers with the provided options
// and merges their returned streams into a single output channel.
//
// This function intentionally operates purely on channels:
//   - Snapshot mode is represented by discoverers returning a channel that closes
//     after emitting the snapshot.
//   - Follow mode is represented by discoverers returning a long-lived channel.
func (mod *Module) DiscoverServices(
	ctx *astral.Context,
	caller *astral.Identity,
	opts services.DiscoverOptions,
) (<-chan services.ServiceChange, error) {
	// Collect source channels.
	streams := make([]<-chan services.ServiceChange, 0, len(mod.discoverers.Clone()))

	for _, discoverer := range mod.discoverers.Clone() {
		s, err := discoverer.DiscoverService(ctx, caller, opts)
		if err != nil {
			mod.log.Logv(1, "discoverer failed: %v", err)
			continue
		}
		if s == nil {
			continue
		}
		streams = append(streams, s)
	}

	return mergeChannels(ctx, streams...), nil
}

func mergeChannels(ctx *astral.Context, channels ...<-chan services.ServiceChange) <-chan services.ServiceChange {
	out := make(chan services.ServiceChange, 16)

	var wg sync.WaitGroup
	for _, c := range channels {
		if c == nil {
			continue
		}

		ch := c
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case msg, ok := <-ch:
					if !ok {
						return
					}
					select {
					case out <- msg:
					case <-ctx.Done():
						return
					}
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
