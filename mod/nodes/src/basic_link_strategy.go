package nodes

import (
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/sig"
)

type BasicLinkStrategy struct {
	mod     *Module
	network string
	target  *astral.Identity
	done    chan struct{}
}

var _ nodes.LinkStrategy = &BasicLinkStrategy{}

type BasicLinkStrategyFactory struct {
	mod     *Module
	network string
}

var _ nodes.StrategyFactory = &BasicLinkStrategyFactory{}

func (f *BasicLinkStrategyFactory) Build(target *astral.Identity) nodes.LinkStrategy {
	return &BasicLinkStrategy{
		mod:     f.mod,
		network: f.network,
		target:  target,
		done:    make(chan struct{}),
	}
}

func (s *BasicLinkStrategy) Done() <-chan struct{} {
	return s.done
}

func (s *BasicLinkStrategy) Signal(ctx *astral.Context) {
	go func() {
		defer close(s.done)

		var wg sync.WaitGroup
		var out sig.Value[*Stream]
		var workers = DefaultWorkerCount

		wctx, cancel := ctx.WithCancel()
		defer cancel()

		resolved, err := s.mod.ResolveEndpoints(ctx, s.target)
		if err != nil {
			return
		}

		endpointsChan := sig.FilterChan(resolved, func(e exonet.Endpoint) bool {
			return e.Network() == s.network
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

					conn, err := s.mod.Exonet.Dial(wctx, endpoint)
					if err != nil {
						continue
					}

					stream, err := s.mod.peers.EstablishOutboundLink(ctx, s.target, conn)
					if err != nil {
						conn.Close()
						continue
					}

					if _, ok := out.Swap(nil, stream); ok {
						cancel()
					} else {
						stream.CloseWithError(nodes.ErrExcessStream)
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
			return
		}

		used := s.mod.linkPool.notifyStreamWatchers(stream)
		if !used {
			stream.CloseWithError(nodes.ErrExcessStream)
		}
	}()
}
