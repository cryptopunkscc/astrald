package linker

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral/link"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/network/graph"
	"github.com/cryptopunkscc/astrald/node/network/peer"
	"sync"
	"time"
)

const activeLinkPeriod = time.Minute

type LinkManager struct {
	wakeCh   chan id.Identity
	localID  id.Identity
	resolver graph.Resolver
	peers    *peer.Set

	active   map[string]context.CancelFunc
	activeMu sync.Mutex
}

func NewManager(localID id.Identity, peer *peer.Set, resolver graph.Resolver) *LinkManager {
	return &LinkManager{
		localID:  localID,
		peers:    peer,
		resolver: resolver,
		active:   make(map[string]context.CancelFunc),
		wakeCh:   make(chan id.Identity, 1),
	}
}

func (m *LinkManager) Run(ctx context.Context) <-chan *link.Link {
	outCh := make(chan *link.Link)

	go func() {
		defer close(outCh)
		for {
			select {
			case nodeID := <-m.wakeCh:
				go m.wake(ctx, nodeID, outCh)
			case <-ctx.Done():
				return
			}
		}
	}()

	return outCh
}

func (m *LinkManager) Wake(remoteId id.Identity) {
	m.wakeCh <- remoteId
}

func (m *LinkManager) Sleep(remoteId id.Identity) {
	m.activeMu.Lock()
	defer m.activeMu.Unlock()

	hex := remoteId.PublicKeyHex()

	cancel, found := m.active[hex]
	if found {
		cancel()
	}
}

func (m *LinkManager) wake(ctx context.Context, remoteId id.Identity, links chan<- *link.Link) {
	m.activeMu.Lock()
	defer m.activeMu.Unlock()

	hex := remoteId.PublicKeyHex()
	if _, found := m.active[hex]; found {
		return
	}

	peer := m.peers.Peer(remoteId)
	linkerCtx, cancel := context.WithTimeout(ctx, activeLinkPeriod)

	m.active[hex] = cancel

	go func() {
		for link := range SustainPeerLink(linkerCtx, m.localID, peer, m.resolver) {
			links <- link
		}
	}()

	go func() {
		<-linkerCtx.Done()
		m.activeMu.Lock()
		defer m.activeMu.Unlock()
		delete(m.active, hex)
	}()
}
