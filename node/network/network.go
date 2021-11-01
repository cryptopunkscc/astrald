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
	"github.com/cryptopunkscc/astrald/node/network/linker"
	"github.com/cryptopunkscc/astrald/node/network/route"
	"github.com/cryptopunkscc/astrald/node/storage"
	"io"
	"sync"
	"time"
)

const defaultLinkIdleTimeout = 10 * time.Minute
const defaultLinkTimeout = 60 * time.Second

type Network struct {
	*View
	Linker *linker.Linker
	Router *route.BasicRouter

	config   Config
	identity id.Identity
	store    storage.Store
	inet     *inet.Inet
	tor      *tor.Tor
}

func NewNetwork(config Config, identity id.Identity, store storage.Store) *Network {
	var err error
	router := route.NewBasicRouter()
	n := &Network{
		View:     NewView(),
		Router:   router,
		Linker:   linker.NewLinker(identity, router),
		config:   config,
		identity: identity,
		store:    store,
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

	return n
}

func (network *Network) Query(ctx context.Context, remoteID id.Identity, query string) (io.ReadWriteCloser, error) {
	network.Linker.Wake(remoteID)

	peer := network.View.Peer(remoteID)

	select {
	case <-peer.WaitLinked():
		return peer.Query(query)
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(defaultLinkTimeout):
		return nil, errors.New("timeout")
	}
}

func (network *Network) Route(onlyPublic bool) *route.Route {
	addrs := make([]infra.Addr, 0)

	for _, a := range astral.Addresses() {
		if onlyPublic && !a.Public {
			continue
		}
		addrs = append(addrs, a.Addr)
	}

	return &route.Route{
		Identity:  network.identity,
		Addresses: addrs,
	}
}

func (network *Network) onLink(ctx context.Context, link *link.Link, reqCh chan<- link.Request, evCh chan<- Event) error {
	err := network.View.addLink(link)
	if err != nil {
		return err
	}

	// forward link's requests
	go func() {
		for req := range link.Requests() {
			reqCh <- req
		}
		evCh <- Event{Type: EventLinkDown, Link: link}
	}()

	// set an idle timeout for the link // TODO: unlinking strategy should not be link-based
	go func() {
		for {
			if err := link.WaitIdle(ctx, defaultLinkIdleTimeout); err != nil {
				return
			}
			link.Close()
		}
	}()

	evCh <- Event{Type: EventLinkUp, Link: link}

	peer := network.Peer(link.RemoteIdentity())
	if peer.LinkCount() == 1 {
		// if it's a new inbound peer, try to link back to improve link quality
		if link.Outbound() == false {
			network.Linker.Wake(peer.Identity())
		}
		evCh <- Event{Type: EventPeerLinked, Peer: peer}
	}

	return nil
}

func (network *Network) loadState() error {
	data, err := network.store.LoadBytes("routes")
	if err != nil {
		return err
	}

	return network.Router.AddPacked(data)
}

func (network *Network) storeState() error {
	return network.store.StoreBytes("routes", network.Router.Pack())
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
