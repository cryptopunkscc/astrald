package services

import (
	"context"
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/services"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opServiceDiscoveryArgs struct {
	Out    string `query:"optional"`
	Follow bool   `query:"optional"`
}

func (mod *Module) OpDiscovery(ctx *astral.Context, q shell.Query, args opServiceDiscoveryArgs) error {
	ch := astral.NewChannelFmt(q.Accept(), "", args.Out)
	defer ch.Close()

	caller := q.Caller()

	snapshotOpts := services.DiscoverOptions{Snapshot: true, Follow: false}
	followOpts := services.DiscoverOptions{Snapshot: false, Follow: args.Follow}

	// Phase 1: snapshot
	for _, discoverer := range mod.discoverers {
		s, err := discoverer.DiscoverService(ctx, caller, snapshotOpts)
		if err != nil {
			mod.log.Logv(1, "discoverer snapshot failed: %v", err)
			continue
		}
		if s == nil {
			continue
		}

		for {
			select {
			case <-ctx.Done():
				return nil
			case update, ok := <-s:
				if !ok {
					goto nextDiscoverer
				}
				if err := ch.Write(&update); err != nil {
					return err
				}
			}
		}
	nextDiscoverer:
	}

	if err := ch.Write(&astral.EOS{}); err != nil {
		return err
	}

	if !args.Follow {
		return nil
	}

	streams := make([]<-chan services.ServiceChange, 0, len(mod.discoverers))
	for _, discoverer := range mod.discoverers {
		s, err := discoverer.DiscoverService(ctx, caller, followOpts)
		if err != nil {
			mod.log.Logv(1, "discoverer follow failed: %v", err)
			continue
		}
		if s != nil {
			streams = append(streams, s)
		}
	}

	merged := mergeChannels(ctx, streams...)
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

// mergeChannels merges multiple ServiceChange channels into a single channel.
func mergeChannels(ctx context.Context, channels ...<-chan services.ServiceChange) <-chan services.ServiceChange {
	out := make(chan services.ServiceChange)
	var wg sync.WaitGroup

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

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
