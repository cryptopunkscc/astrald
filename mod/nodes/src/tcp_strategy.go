package nodes

import (
	"errors"
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/sig"
)

type TCPStrategy struct {
	mod      *Module
	target   *astral.Identity
	produced chan<- *Stream
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

func (s *TCPStrategy) Activate(ctx *astral.Context, endpoints []exonet.Endpoint) error {
	if len(endpoints) == 0 {
		return errors.New("no endpoints")
	}

	var wg sync.WaitGroup
	var out sig.Value[*Stream]

	wctx, cancel := ctx.WithCancel()
	defer cancel()

	var started bool
	for _, ep := range endpoints {
		if err := s.inFlight.Add(ep.Address()); err != nil {
			continue
		}

		started = true
		wg.Add(1)
		go func(endpoint exonet.Endpoint) {
			defer wg.Done()
			defer s.inFlight.Remove(endpoint.Address())

			select {
			case <-wctx.Done():
				return
			default:
			}

			stream, err := s.mod.peers.connectAt(wctx, s.target, endpoint)
			if err != nil {
				return
			}

			if _, ok := out.Swap(nil, stream); ok {
				cancel()
			} else {
				stream.CloseWithError(errors.New("excess stream"))
			}
		}(ep)
	}

	if !started {
		return errors.New("no endpoints")
	}

	go func() {
		wg.Wait()
		cancel()
	}()

	<-wctx.Done()

	stream := out.Get()
	if stream == nil {
		return errors.New("no endpoint could be reached")
	}

	select {
	case s.produced <- stream:
	default:
	}

	return nil
}
