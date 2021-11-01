package linker

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral/link"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/network/route"
	"sync"
)

type Linker struct {
	wake    chan id.Identity
	localID id.Identity
	router  route.Router

	active   map[string]Strategy
	activeMu sync.Mutex
}

func NewLinker(localID id.Identity, router route.Router) *Linker {
	return &Linker{
		localID: localID,
		router:  router,
		active:  make(map[string]Strategy),
		wake:    make(chan id.Identity, 1),
	}
}

func (linker *Linker) Run(ctx context.Context) <-chan *link.Link {
	outCh := make(chan *link.Link)

	go func() {
		defer close(outCh)
		for {
			select {
			case nodeID := <-linker.wake:
				go linker.realWake(ctx, nodeID, outCh)
			case <-ctx.Done():
				return
			}
		}
	}()

	return outCh
}

func (linker *Linker) Wake(remoteId id.Identity) {
	linker.wake <- remoteId
}

func (linker *Linker) realWake(ctx context.Context, remoteId id.Identity, links chan<- *link.Link) {
	linker.activeMu.Lock()
	defer linker.activeMu.Unlock()

	hex := remoteId.PublicKeyHex()

	if _, found := linker.active[hex]; !found {
		linker.active[hex] = RunDefaultStrategy(ctx, linker.localID, remoteId, linker.router)

		go func() {
			for link := range linker.active[hex].Links() {
				links <- link
			}
		}()
	}

	linker.active[hex].Wake()
}
