package nodes

import (
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/sig"
)

// BasicLinkStrategy dials all known endpoints of the target node in parallel
// and uses the first successful connection to establish a link.
type BasicLinkStrategy struct {
	mod      *Module
	network  string
	target   *astral.Identity
	inflight sig.Set[string] // tracks endpoints currently being dialed

	mu         sync.Mutex
	activeDone chan struct{}
}

var _ nodes.LinkStrategy = &BasicLinkStrategy{}

func (s *BasicLinkStrategy) Signal(ctx *astral.Context) {
	s.mu.Lock()
	if s.activeDone != nil {
		s.mu.Unlock()
		return
	}
	s.activeDone = make(chan struct{})
	s.mu.Unlock()

	go func() {
		defer func() {
			s.mu.Lock()
			close(s.activeDone)
			s.activeDone = nil
			s.mu.Unlock()
		}()

		var wg sync.WaitGroup
		var winner sig.Value[*Stream]

		wctx, cancel := ctx.WithCancel()
		defer cancel()

		resolved, err := s.mod.ResolveEndpoints(ctx, s.target)
		if err != nil {
			s.mod.log.Log("resolve endpoints failed: %v", err)
			return
		}

		endpoints := sig.FilterChan(resolved, func(e exonet.Endpoint) bool {
			return e.Network() == s.network
		})

		wg.Add(DefaultWorkerCount)
		for i := 0; i < DefaultWorkerCount; i++ {
			go func() {
				defer wg.Done()
				for {
					select {
					case <-wctx.Done():
						return
					case endpoint, ok := <-endpoints:
						if !ok {
							return
						}

						stream := s.tryEndpoint(wctx, endpoint)
						if stream != nil {
							if _, ok := winner.Swap(nil, stream); ok {
								cancel()
							} else {
								stream.CloseWithError(nodes.ErrExcessStream)
							}

							return
						}
					}
				}
			}()
		}

		wg.Wait()

		stream := winner.Get()
		if stream == nil {
			return
		}

		if !s.mod.linkPool.notifyStreamWatchers(stream) {
			stream.CloseWithError(nodes.ErrExcessStream)
		}
	}()
}

func (s *BasicLinkStrategy) Done() <-chan struct{} {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.activeDone == nil {
		ch := make(chan struct{})
		close(ch)
		return ch
	}
	return s.activeDone
}

func (s *BasicLinkStrategy) tryEndpoint(ctx *astral.Context, endpoint exonet.Endpoint) *Stream {
	addr := endpoint.Address()
	if err := s.inflight.Add(addr); err != nil {
		return nil
	}
	defer s.inflight.Remove(addr)

	conn, err := s.mod.Exonet.Dial(ctx, endpoint)
	if err != nil {
		return nil
	}

	stream, err := s.mod.peers.EstablishOutboundLink(ctx, s.target, conn)
	if err != nil {
		conn.Close()
		return nil
	}

	return stream
}

// factory

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
	}
}
