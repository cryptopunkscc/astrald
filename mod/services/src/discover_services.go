package services

import (
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/services"
)

// DiscoverServices runs all registered ServiceDiscoverers with the provided options
// and merges their returned streams into a single output channel.
//
// Stream semantics:
//   - If opts.Snapshot is true, discoverers should emit ServiceChangeTypeSnapshot entries.
//   - When a discoverer has finished its snapshot phase, it should emit ServiceChangeTypeFlush.
//     (This is treated as a control message and is not forwarded.)
//   - If opts.Follow is true, discoverers should then emit ServiceChangeTypeUpdate entries.
//
// This function forwards snapshot + update entries as they arrive (discoverers run concurrently).
// When all discoverers complete snapshotting, this function emits a single *astral.EOS to mark
// the snapshot boundary, then continues with updates.
func (mod *Module) DiscoverServices(
	ctx *astral.Context,
	caller *astral.Identity,
	opts services.DiscoverOptions,
) (<-chan astral.Object, error) {
	out := make(chan services.ServiceChange, 128)
	wg := &sync.WaitGroup{}

	var discoverers = mod.discoverers.Clone()
	for _, discoverer := range discoverers {

		wg.Add(1)
		// each discoverer runs at once in goroutine
		go func() {
			defer wg.Done()
			s, err := discoverer.DiscoverService(ctx, caller, opts)
			if err != nil {
				mod.log.Logv(1, "discoverer failed: %v", err)
			}

			streams = append(streams, s)

		snapshotLoop:
			for serviceChange := range s {
				if serviceChange.Type == services.ServiceChangeTypeFlush {
					break snapshotLoop
				}

				// forward serviceChanges to out channel
				select {
				case <-ctx.Done():
					return
				case out <- serviceChange:
				}
			}
		}()
	}

	wg.Wait()
	// at this point we have collected all snapshot entries from discoverers
	// We should send EOS and then

	for _, s := range streams {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for serviceChange := range s {
				select {
				case <-ctx.Done():
					return
				case out <- serviceChange:
				}
			}
		}()
	}

	return nil
}
