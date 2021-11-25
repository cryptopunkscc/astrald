package network

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/link"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/infra/inet"
	"github.com/cryptopunkscc/astrald/infra/tor"
	"github.com/cryptopunkscc/astrald/node/network/graph"
	"github.com/cryptopunkscc/astrald/node/network/linker"
	_peer "github.com/cryptopunkscc/astrald/node/network/peer"
	"github.com/cryptopunkscc/astrald/node/storage"
	sync2 "github.com/cryptopunkscc/astrald/sync"
	"io"
	"sync"
	"time"
)

const defaultIdleTimeout = 15 * time.Minute
const defaultLinkTimeout = time.Minute

type Network struct {
	*_peer.Set
	Linker *linker.LinkManager
	Graph  *graph.Graph

	config  Config
	localID id.Identity
	store   storage.Store
	inet    *inet.Inet
	tor     *tor.Tor
}

func NewNetwork(config Config, identity id.Identity, store storage.Store) *Network {
	var err error

	n := &Network{
		Set:     _peer.NewSet(),
		config:  config,
		localID: identity,
		store:   store,
	}

	// Configure internet
	n.inet = inet.New(config.Inet)

	err = astral.AddNetwork(n.inet)
	if err != nil {
		panic(err)
	}

	// Configure tor
	n.tor = tor.New(config.Tor)
	err = astral.AddNetwork(n.tor)
	if err != nil {
		panic(err)
	}

	// graph needs to be set up after networks so that all address parsers are loaded
	n.Graph = graph.New(store)
	n.Linker = linker.NewManager(identity, n.Set, n.Graph)

	return n
}

func (network *Network) Query(ctx context.Context, remoteID id.Identity, query string) (io.ReadWriteCloser, error) {
	network.Linker.Wake(remoteID)

	remotePeer := network.Set.Peer(remoteID)

	// Wait for peer to be linked
	waitCtx, _ := context.WithTimeout(ctx, defaultLinkTimeout)

	<-_peer.LinkedGate(waitCtx, remotePeer).Wait()

	l := link.Select(remotePeer.Links.Links(), link.Fastest)
	if l == nil {
		return nil, errors.New("no link available")
	}

	return l.Query(query)
}

func (network *Network) Alias() string {
	return network.config.Alias
}

func (network *Network) Info(onlyPublic bool) *graph.Info {
	addrs := make([]infra.Addr, 0)

	for _, a := range astral.Addresses() {
		if onlyPublic && !a.Public {
			continue
		}
		addrs = append(addrs, a.Addr)
	}

	info := &graph.Info{
		Identity:  network.localID,
		Addresses: addrs,
	}

	if !onlyPublic {
		info.Alias = network.Alias()
	}

	return info
}

func (network *Network) onLink(ctx context.Context, link *link.Link, reqCh chan<- link.Request, evCh chan<- Event) error {
	err := network.Set.AddLink(link)
	if err != nil {
		return err
	}

	peer := network.Peer(link.RemoteIdentity())

	// forward link's requests
	go func() {
		for req := range link.Requests() {
			reqCh <- req
		}
		evCh <- Event{Type: EventLinkDown, Link: link}

		if len(peer.Links.Links()) == 0 {
			evCh <- Event{Type: EventPeerUnlinked, Peer: peer}
		}
	}()

	evCh <- Event{Type: EventLinkUp, Link: link}

	if len(peer.Links.Links()) == 1 {
		// if it's a new inbound peer, try to link back to improve link quality
		if link.Outbound() == false {
			network.Linker.Wake(peer.Identity())
		}
		evCh <- Event{Type: EventPeerLinked, Peer: peer}

		sync2.On(ctx, sync2.Timeout(ctx, peer, defaultIdleTimeout), func() {
			for l := range peer.Links.Links() {
				l.Close()
			}
		})
	}

	return nil
}

func mergeLinkChans(chans ...<-chan *link.Link) <-chan *link.Link {
	outCh := make(chan *link.Link)

	var wg sync.WaitGroup

	for _, ch := range chans {
		ch := ch
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range ch {
				outCh <- i
			}
		}()
	}

	go func() {
		wg.Wait()
		close(outCh)
	}()

	return outCh
}
