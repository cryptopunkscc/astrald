package linker

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral/link"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/route"
	"sync"
)

type Linker struct {
	ctx     context.Context
	localID id.Identity
	router  route.Router

	links    chan *link.Link
	active   map[string]Strategy
	activeMu sync.Mutex
}

func NewLinker(ctx context.Context, localID id.Identity, router route.Router) *Linker {
	return &Linker{
		ctx:     ctx,
		localID: localID,
		router:  router,
		links:   make(chan *link.Link),
		active:  make(map[string]Strategy),
	}
}

func (linker *Linker) Wake(remoteId id.Identity) {
	linker.activeMu.Lock()
	defer linker.activeMu.Unlock()

	hex := remoteId.PublicKeyHex()

	if _, found := linker.active[hex]; !found {
		linker.active[hex] = RunDefaultStrategy(linker.ctx, linker.localID, remoteId, linker.router)

		go func() {
			for link := range linker.active[hex].Links() {
				linker.links <- link
			}
		}()
	}

	linker.active[hex].Wake()
}

func (linker *Linker) Links() <-chan *link.Link {
	return linker.links
}
