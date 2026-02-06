package nodes

import (
	"errors"
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/tcp"
	"github.com/cryptopunkscc/astrald/sig"
)

type TCPStrategy struct {
	mod      *Module
	produced chan<- *Stream
	target   *astral.Identity

	inFlight sig.Set[string]
}

var _ LinkStrategy = &TCPStrategy{}

// fixme: simplify & improve
func NewTCPStrategy(mod *Module, target *astral.Identity, produced chan<- *Stream) *TCPStrategy {
	return &TCPStrategy{
		mod:      mod,
		target:   target,
		produced: produced,
	}
}

func (s *TCPStrategy) Activate(ctx *astral.Context) chan error {
	done := make(chan error, 1)

	go func() {
		var wg sync.WaitGroup
		var out sig.Value[*Stream]
		var workers = DefaultWorkerCount

		wctx, cancel := ctx.WithCancel()
		defer cancel()

		resolved, err := s.mod.ResolveEndpoints(ctx, s.target)
		if err != nil {
			done <- err
			return
		}

		endpointsChan := sig.FilterChan(resolved, func(e exonet.Endpoint) bool {
			return e.Network() == tcp.ModuleName
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
			done <- nodes.ErrNoEndpointReached
			return
		}

		select {
		case s.produced <- stream:
		default:
		}

		done <- nil
	}()

	return done
}
