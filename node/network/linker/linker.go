package linker

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral/link"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/network/peer"
	"github.com/cryptopunkscc/astrald/node/network/route"
	"sync"
)

type Linker struct {
	wakeCh  chan id.Identity
	localID id.Identity
	router  route.Router
	peers   *peer.Pool

	active   map[string]context.CancelFunc
	activeMu sync.Mutex
}

func NewLinker(localID id.Identity, peer *peer.Pool, router route.Router) *Linker {
	return &Linker{
		localID: localID,
		peers:   peer,
		router:  router,
		active:  make(map[string]context.CancelFunc),
		wakeCh:  make(chan id.Identity, 1),
	}
}

func (linker *Linker) Run(ctx context.Context) <-chan *link.Link {
	outCh := make(chan *link.Link)

	go func() {
		defer close(outCh)
		for {
			select {
			case nodeID := <-linker.wakeCh:
				go linker.realWake(ctx, nodeID, outCh)
			case <-ctx.Done():
				return
			}
		}
	}()

	return outCh
}

func (linker *Linker) Wake(remoteId id.Identity) {
	linker.wakeCh <- remoteId
}

func (linker *Linker) Sleep(remoteId id.Identity) {
	linker.activeMu.Lock()
	defer linker.activeMu.Unlock()

	hex := remoteId.PublicKeyHex()

	cancel, found := linker.active[hex]
	if found {
		cancel()
	}
}

func (linker *Linker) realWake(ctx context.Context, remoteId id.Identity, links chan<- *link.Link) {
	linker.activeMu.Lock()
	defer linker.activeMu.Unlock()

	hex := remoteId.PublicKeyHex()
	if _, found := linker.active[hex]; found {
		return
	}

	peer := linker.peers.Peer(remoteId)
	linkerCtx, cancel := context.WithTimeout(ctx, activeLinkPeriod)

	linker.active[hex] = cancel

	go func() {
		for link := range KeepLinked(linkerCtx, linker.localID, peer, linker.router) {
			links <- link
		}
	}()

	go func() {
		<-linkerCtx.Done()
		linker.activeMu.Lock()
		defer linker.activeMu.Unlock()
		delete(linker.active, hex)
	}()
}
