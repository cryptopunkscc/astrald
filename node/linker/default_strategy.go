package linker

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/link"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/route"
	"sync"
	"time"
)

type DefaultStrategy struct {
	localID  id.Identity
	remoteID id.Identity
	router   route.Router
}

func NewDefaultStrategy(localID id.Identity, remoteID id.Identity, router route.Router) *DefaultStrategy {
	return &DefaultStrategy{localID: localID, remoteID: remoteID, router: router}
}

func (s *DefaultStrategy) runNetwork(ctx context.Context, netName string) <-chan *link.Link {
	outCh := make(chan *link.Link)

	go func() {
		defer close(outCh)

		for {
			r := s.router.Route(s.remoteID)
			if r == nil {
				time.Sleep(5 * time.Second)
				continue
			}

			for _, addr := range r.Addresses {
				if addr.Network() != netName {
					continue
				}

				lnk, err := astral.Link(s.localID, s.remoteID, addr)
				if err != nil {
					// can be very spammy
					// log.Println(err)
					continue
				}
				outCh <- lnk
				<-lnk.WaitClose()
				break
			}

			time.Sleep(5 * time.Second)
		}
	}()

	return outCh
}

func (s *DefaultStrategy) Run(ctx context.Context) <-chan *link.Link {
	outCh := make(chan *link.Link)

	go func() {
		defer close(outCh)

		var wg sync.WaitGroup

		for _, netName := range astral.NetworkNames() {
			wg.Add(1)
			go func(netName string) {
				defer wg.Done()
				links := s.runNetwork(ctx, netName)
				for lnk := range links {
					outCh <- lnk
				}
			}(netName)
		}

		wg.Wait()
	}()

	return outCh
}
