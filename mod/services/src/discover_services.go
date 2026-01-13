package services

import (
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/services"
)

// DiscoverServices runs all registered ServiceDiscoverers with the provided options and merges
// their returned streams into a single output channel.
func (mod *Module) DiscoverServices(
	ctx *astral.Context,
	caller *astral.Identity,
	opts services.DiscoverOptions,
) (snapshot []services.ServiceChange, updates <-chan services.ServiceChange, err error) {
	wg := &sync.WaitGroup{}
	streams := make([]<-chan services.ServiceChange, 0)

	var discoverers = mod.discoverers.Clone()
	for _, d := range discoverers {
		var discoverer = d
		wg.Add(1)
		// each discoverer runs at once in goroutine
		go func() {
			defer wg.Done()
			s, err := discoverer.DiscoverService(ctx, caller, opts)
			if err != nil {
				// NOTE: should it fail the whole discovery if one discoverer fails?
				mod.log.Logv(1, "discoverer failed: %v", err)
				return
			}

		snapshotLoop:
			for {
				select {
				case <-ctx.Done():
					return
				case serviceChange, ok := <-s:
					if !ok {
						return
					}

					if serviceChange.Type == services.ServiceChangeTypeFlush {
						streams = append(streams, s)
						break snapshotLoop
					}

					snapshot = append(snapshot, serviceChange)
				}
			}
		}()
	}

	wg.Wait()
	// snapshot collected

	return snapshot, mergeDiscoverStreams(streams), nil
}

func mergeDiscoverStreams(streams []<-chan services.ServiceChange) <-chan services.ServiceChange {
	return nil
}
