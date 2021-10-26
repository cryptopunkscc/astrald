package linker

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral/link"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/route"
	"sync"
)

type Linker struct {
	ctx      context.Context
	localID  id.Identity
	router   route.Router
	active   map[string]Strategy
	activeMu sync.Mutex
	links    chan *link.Link
}

func NewLinker(localID id.Identity, router route.Router) *Linker {
	return &Linker{
		localID: localID,
		router:  router,
		links:   make(chan *link.Link),
		active:  make(map[string]Strategy),
	}
}

func (linker *Linker) Run(ctx context.Context) {
	linker.ctx = ctx
}

func (linker *Linker) Wake(remoteId id.Identity) {
	linker.activeMu.Lock()
	defer linker.activeMu.Unlock()

	hex := remoteId.PublicKeyHex()
	s, found := linker.active[hex]
	if found {
		return
	}

	s = NewDefaultStrategy(linker.localID, remoteId, linker.router)
	linker.active[hex] = s

	go func() {
		for l := range s.Run(linker.ctx) {
			linker.links <- l
		}
	}()
}

func (linker *Linker) Links() <-chan *link.Link {
	return linker.links
}

func (linker *Linker) getStrategy(remoteId id.Identity) Strategy {
	linker.activeMu.Lock()
	defer linker.activeMu.Unlock()

	hex := remoteId.PublicKeyHex()
	s, found := linker.active[hex]
	if found {
		return s
	}
	s = NewDefaultStrategy(linker.localID, remoteId, linker.router)
	linker.active[hex] = s
	return s
}
