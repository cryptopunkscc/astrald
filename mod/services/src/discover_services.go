package services

import (
	"sync"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/services"
	"github.com/cryptopunkscc/astrald/sig"
)

// DiscoverServices runs all registered ServiceDiscoverers with the provided options and merges
// their returned streams into a single output channel.
func (mod *Module) DiscoverServices(
	ctx *astral.Context,
	caller *astral.Identity,
	opts services.DiscoverOptions,
) (snapshot []services.ServiceChange, updates <-chan services.ServiceDiscoveryResult, err error) {
	discoverers := mod.discoverers.Clone()

	phasedStreamsCh := make(chan *sig.PhaseSplitter[services.ServiceDiscoveryResult], len(discoverers))
	var wg = &sync.WaitGroup{}

	for _, d := range discoverers {
		discoverer := d

		wg.Add(1)
		go func() {
			defer wg.Done()
			results, err := discoverer.DiscoverService(ctx, caller, opts)
			if err != nil {
				mod.log.Logv(1, "discoverer failed: %v", err)
				return
			}

			phasedStream := sig.NewPhaseSplitter(
				ctx,
				results,
				services.IsDiscoveryFlush,
			)

			phasedStreamsCh <- phasedStream
		}()
	}

	wg.Wait()
	close(phasedStreamsCh)

	phasedStreams := sig.ChanToArray(phasedStreamsCh)

	beforeStreams := make([]<-chan services.ServiceDiscoveryResult, 0, len(phasedStreams))
	for _, ps := range phasedStreams {
		beforeStreams = append(beforeStreams, ps.Before())
	}

	snapshotCtx, cancelSnapshot := ctx.WithTimeout(30 * time.Second)
	defer cancelSnapshot()

	snapshotEvents := sig.ChanCollectAll(snapshotCtx, beforeStreams...)
	for _, ev := range snapshotEvents {
		if ev.Kind == services.DiscoveryEventChange {
			snapshot = append(snapshot, ev.Change)
		}
	}

	updateStreams := make([]<-chan services.ServiceDiscoveryResult, 0, len(phasedStreams))
	for _, ps := range phasedStreams {
		updateStreams = append(updateStreams, ps.After())
	}

	return snapshot, sig.ChanFanIn(ctx, updateStreams...), nil
}
