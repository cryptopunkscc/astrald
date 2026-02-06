package nodes

import (
	"errors"
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/tor"
	"github.com/cryptopunkscc/astrald/sig"
)

type TorStrategy struct {
	mod      *Module
	produced chan<- *Stream
	target   *astral.Identity

	inFlight sig.Set[string]
}

var _ LinkStrategy = &TorStrategy{}

func NewTorStrategy(mod *Module, target *astral.Identity, produced chan<- *Stream) *TorStrategy {
	return &TorStrategy{
		mod:      mod,
		target:   target,
		produced: produced,
	}
}

func (s *TorStrategy) Activate(ctx *astral.Context) error {
	var wg sync.WaitGroup
	var out sig.Value[*Stream]
	var workers = DefaultWorkerCount

	wctx, cancel := ctx.WithCancel()
	defer cancel()

	resolved, err := s.mod.ResolveEndpoints(ctx, s.target)
	if err != nil {
		return err
	}

	endpointsChan := sig.FilterChan(resolved, func(e exonet.Endpoint) bool {
		return e.Network() == tor.ModuleName
	})

	wg.Add(workers)
	for range workers {
		go func() {
			defer wg.Done()
			for {
				var endpoint exonet.Endpoint
				var ok bool

				select {
				case <-wctx.Done():
					return
				case endpoint, ok = <-endpointsChan:
					if !ok {
						return
					}
				}

				if err := s.inFlight.Add(endpoint.Address()); err != nil {
					continue
				}

				stream, err := s.mod.peers.connectAt(wctx, s.target, endpoint)
				s.inFlight.Remove(endpoint.Address())
				if err != nil {
					continue
				}

				if _, ok := out.Swap(nil, stream); ok {
					cancel()
				} else {
					stream.CloseWithError(errors.New("excess stream"))
				}

				return
			}
		}()
	}

	go func() {
		wg.Wait()
		cancel()
	}()

	<-wctx.Done()

	stream := out.Get()
	if stream == nil {
		return nodes.ErrNoEndpointReached
	}

	select {
	case s.produced <- stream:
	default:
	}

	return nil
}
