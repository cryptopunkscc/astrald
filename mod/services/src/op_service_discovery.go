package services

import (
	"context"
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/services"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opServiceDiscoveryArgs struct {
	Out string `query:"optional"`
}

func (mod *Module) OpServiceDiscovery(ctx *astral.Context, q shell.Query, args opServiceDiscoveryArgs) error {
	ch := astral.NewChannelFmt(q.Accept(), "", args.Out)
	defer ch.Close()

	// Get the caller's identity
	caller := q.Caller()

	// Collect snapshots and channels from all discoverers
	var allSnapshots []*services.Service
	var channels []<-chan services.ServiceChange

	for _, discoverer := range mod.discoverers {
		snapshot, discoverCh, err := discoverer.DiscoverService(ctx, caller)
		if err != nil {
			mod.log.Logv(1, "discoverer failed: %v", err)
			continue
		}

		if snapshot != nil {
			allSnapshots = append(allSnapshots, snapshot)
		}

		channels = append(channels, discoverCh)
	}

	// Send all snapshots first (only non-nil services)
	for _, svc := range allSnapshots {
		change := services.ServiceChange{
			Enabled: true,
			Service: *svc,
		}
		if err := ch.Write(&change); err != nil {
			return err
		}
	}

	// Send EOS to mark end of snapshot
	if err := ch.Write(&astral.EOS{}); err != nil {
		return err
	}

	// Merge all discoverer channels for live updates
	merged := mergeChannels(ctx, channels...)

	// Stream live updates to client
	for {
		select {
		case <-ctx.Done():
			return nil
		case update, ok := <-merged:
			if !ok {
				return nil
			}
			if err := ch.Write(&update); err != nil {
				return err
			}
		}
	}
}

// mergeChannels merges multiple ServiceChange channels into a single channel
func mergeChannels(ctx context.Context, channels ...<-chan services.ServiceChange) <-chan services.ServiceChange {
	out := make(chan services.ServiceChange)
	var wg sync.WaitGroup

	// Start a goroutine for each input channel
	for _, c := range channels {
		if c == nil {
			continue
		}
		wg.Add(1)
		go func(ch <-chan services.ServiceChange) {
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
		}(c)
	}

	// Close output channel when all input channels are done
	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
