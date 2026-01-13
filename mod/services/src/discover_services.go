package services

import (
	"context"
	"sync"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/services"
)

type discoverEventKind uint8

const (
	eventSnapshot discoverEventKind = iota
	eventFlush
)

type discoverResult struct {
	kind   discoverEventKind
	change services.ServiceChange
	stream <-chan services.ServiceChange
}

// DiscoverServices runs all registered ServiceDiscoverers with the provided options and merges
// their returned streams into a single output channel.
func (mod *Module) DiscoverServices(
	ctx *astral.Context,
	caller *astral.Identity,
	opts services.DiscoverOptions,
) (snapshot []services.ServiceChange, updates <-chan services.ServiceChange, err error) {
	snapshotCtx, cancelSnapshot := ctx.WithTimeout(30 * time.Second)
	defer cancelSnapshot()

	wg := &sync.WaitGroup{}

	var discoverers = mod.discoverers.Clone()
	events := make(chan discoverResult, len(discoverers)*4)

	for _, d := range discoverers {
		discoverer := d
		wg.Add(1)
		// Each discoverer runs independently and only emits events.
		go func() {
			defer wg.Done()

			s, err := discoverer.DiscoverService(snapshotCtx, caller, opts)
			if err != nil {
				// NOTE: should it fail the whole discovery if one discoverer fails?
				mod.log.Logv(1, "discoverer failed: %v", err)
				return
			}

			for {
				select {
				case <-snapshotCtx.Done():
					return
				case sc, ok := <-s:
					if !ok {
						return
					}

					serviceChange := sc

					if serviceChange.Type == services.ServiceChangeTypeFlush {
						select {
						case events <- discoverResult{kind: eventFlush, stream: s}:
						case <-snapshotCtx.Done():
						}
						return
					}

					select {
					case <-snapshotCtx.Done():
						return
					case events <- discoverResult{kind: eventSnapshot, change: serviceChange}:
					}
				}
			}
		}()
	}

	// Close events only after all workers are done.
	go func() {
		wg.Wait()
		close(events)
	}()

	streams := make([]<-chan services.ServiceChange, 0, len(discoverers))

	for ev := range events {
		switch ev.kind {
		case eventSnapshot:
			snapshot = append(snapshot, ev.change)
		case eventFlush:
			streams = append(streams, ev.stream)
		}
	}

	return snapshot, mergeDiscoverStreams(ctx, streams), nil
}

func mergeDiscoverStreams(
	ctx context.Context,
	streams []<-chan services.ServiceChange,
) <-chan services.ServiceChange {
	out := make(chan services.ServiceChange, len(streams)*2)

	if len(streams) == 0 {
		close(out)
		return out
	}

	var wg sync.WaitGroup
	wg.Add(len(streams))

	for _, s := range streams {
		stream := s
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case v, ok := <-stream:
					if !ok {
						return
					}
					select {
					case out <- v:
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
