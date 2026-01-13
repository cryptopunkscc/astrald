package services

import (
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
) (snapshot []services.ServiceChange, updates <-chan services.ServiceChange, err error) {
	// Snapshot should be time-bounded; follow must not be tied to this timeout.
	snapshotCtx, cancelSnapshot := ctx.WithTimeout(30 * time.Second)
	defer cancelSnapshot()

	discoverers := mod.discoverers.Clone()

	phased := make([]*sig.PhasedStream[services.ServiceChange], 0, len(discoverers))

	// 1) Start discoverers and wrap streams.
	for _, d := range discoverers {
		discoverer := d

		discoverCtx := ctx
		if !opts.Follow {
			// Snapshot-only requests should still be time-bounded.
			discoverCtx = snapshotCtx
		}

		s, err := discoverer.DiscoverService(discoverCtx, caller, opts)
		if err != nil {
			mod.log.Logv(1, "discoverer failed: %v", err)
			continue
		}

		ps := sig.NewPhasedStream(
			discoverCtx,
			s,
			func(sc services.ServiceChange) bool {
				return sc.Type == services.ServiceChangeTypeFlush
			},
		)

		phased = append(phased, ps)
	}

	// 2) Collect snapshot phase (Before) with the snapshot timeout.
	beforeStreams := make([]<-chan services.ServiceChange, 0, len(phased))
	for _, ps := range phased {
		beforeStreams = append(beforeStreams, ps.Before())
	}
	snapshot = sig.ChanCollectAll(snapshotCtx, beforeStreams...)

	updateStreams := make([]<-chan services.ServiceChange, 0, len(phased))
	for _, ps := range phased {
		updateStreams = append(updateStreams, ps.After())
	}

	return snapshot, sig.ChanFanIn(ctx, updateStreams...), nil
}
