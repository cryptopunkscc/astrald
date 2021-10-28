package linker

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/link"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/route"
	"sync"
)

type DefaultStrategy struct {
	context  context.Context
	localID  id.Identity
	remoteID id.Identity
	router   route.Router
	wakeCh   chan struct{}
	wakeMu   sync.Mutex
	links    chan *link.Link
}

func RunDefaultStrategy(ctx context.Context, localID id.Identity, remoteID id.Identity, router route.Router) *DefaultStrategy {
	linker := &DefaultStrategy{
		context:  ctx,
		localID:  localID,
		remoteID: remoteID,
		router:   router,
		wakeCh:   make(chan struct{}),
		links:    make(chan *link.Link),
	}
	go linker.run(ctx)
	return linker
}

func (str *DefaultStrategy) Wake() {
	str.wakeMu.Lock()
	defer str.wakeMu.Unlock()

	close(str.wakeCh)
	str.wakeCh = make(chan struct{})
}

func (str *DefaultStrategy) Links() <-chan *link.Link {
	return str.links
}

func (str *DefaultStrategy) run(ctx context.Context) error {
	defer close(str.links)

	var wg sync.WaitGroup

	for _, networkName := range astral.NetworkNames() {
		wg.Add(1)
		go func(networkName string) {
			_ = str.runNetwork(ctx, networkName)
			wg.Done()
		}(networkName)
	}

	wg.Wait()

	return nil
}

func (str *DefaultStrategy) runNetwork(ctx context.Context, networkName string) error {
	for {
		// Get current route for the node
		route := str.router.Route(str.remoteID)
		if route == nil {
			continue
		}

		for _, addr := range route.Addresses {
			if addr.Network() != networkName {
				continue
			}

			link, err := astral.Link(str.localID, str.remoteID, addr)
			if err != nil {
				continue
			}

			str.links <- link
			<-link.WaitClose()
			break
		}

		// Wait to be woken up or canceled
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-str.waitWake():
		}
	}
}

func (str *DefaultStrategy) waitWake() <-chan struct{} {
	str.wakeMu.Lock()
	defer str.wakeMu.Unlock()

	return str.wakeCh
}
