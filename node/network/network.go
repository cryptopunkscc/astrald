package network

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
	iastral "github.com/cryptopunkscc/astrald/infra/astral"
	"github.com/cryptopunkscc/astrald/infra/inet"
	"github.com/cryptopunkscc/astrald/infra/tor"
	"github.com/cryptopunkscc/astrald/logfmt"
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/node/network/contacts"
	"github.com/cryptopunkscc/astrald/node/network/linker"
	"github.com/cryptopunkscc/astrald/node/network/peer"
	"github.com/cryptopunkscc/astrald/node/storage"
	"github.com/cryptopunkscc/astrald/sig"
	"log"
	"sync"
	"time"
)

const defaultPeerIdleTimeout = 15 * time.Minute
const defaultQueryTimeout = time.Minute

type Network struct {
	Contacts *contacts.Contacts

	peers   map[string]*peer.Peer
	peersMu sync.Mutex
	config  Config
	localID id.Identity
	store   storage.Store
	inet    *inet.Inet
	tor     *tor.Tor
	astral  *iastral.Astral

	newLinks chan *link.Link
	Conns    chan infra.Conn

	linkerMu map[*peer.Peer]*sync.Mutex
}

func (n *Network) Identity() id.Identity {
	return n.localID
}

func NewNetwork(config Config, identity id.Identity, store storage.Store) *Network {
	var err error

	n := &Network{
		peers:    make(map[string]*peer.Peer),
		config:   config,
		localID:  identity,
		store:    store,
		linkerMu: make(map[*peer.Peer]*sync.Mutex),
		newLinks: make(chan *link.Link, 1),
		Conns:    make(chan infra.Conn, 1),
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

	// Configure astral for mesh links
	n.astral = iastral.NewAstral(NewAdapter(n), n.config.Astral)
	err = astral.AddNetwork(n.astral)
	if err != nil {
		panic(err)
	}
	// contacts need to be set up after networks so that all address parsers are loaded
	n.Contacts = contacts.New(store)

	return n
}

func (n *Network) Peers() <-chan *peer.Peer {
	n.peersMu.Lock()
	defer n.peersMu.Unlock()

	ch := make(chan *peer.Peer, len(n.peers))
	for _, p := range n.peers {
		ch <- p
	}
	close(ch)
	return ch
}

func (n *Network) Peer(id id.Identity) *peer.Peer {
	n.peersMu.Lock()
	defer n.peersMu.Unlock()

	str := id.String()

	if p, ok := n.peers[str]; ok {
		return p
	}

	p := peer.New(id)
	n.peers[str] = p
	return p
}

func (n *Network) Query(parent context.Context, remoteID id.Identity, query string) (*link.Conn, error) {
	p := n.Peer(remoteID)

	// set up a query context
	ctx, _ := context.WithTimeout(parent, defaultQueryTimeout)

	l, err := n.Connect(ctx, p)
	if err != nil {
		return nil, err
	}

	log.Printf("-> [%s]:%s (%s)\n", logfmt.ID(remoteID), query, l.Network())

	return l.Query(query)
}

func (n *Network) Alias() string {
	return n.config.Alias
}

func (n *Network) Info(onlyPublic bool) *contacts.Info {
	addrs := make([]infra.Addr, 0)

	for _, a := range astral.Addresses() {
		if onlyPublic && !a.Public {
			continue
		}
		addrs = append(addrs, a.Addr)
	}

	info := &contacts.Info{
		Identity:  n.localID,
		Addresses: addrs,
	}

	if !onlyPublic {
		info.Alias = n.Alias()
	}

	return info
}

func (n *Network) onLink(ctx context.Context, link *link.Link, queryCh chan<- *link.Query, evCh chan<- Event) error {
	peer := n.Peer(link.RemoteIdentity())

	if err := peer.Add(link); err != nil {
		return err
	}

	// forward link's requests
	go func() {
		for query := range link.Queries() {
			queryCh <- query
		}
		evCh <- Event{Type: EventLinkDown, Link: link}

		if len(peer.Links()) == 0 {
			evCh <- Event{Type: EventPeerUnlinked, Peer: peer}
		}
	}()

	evCh <- Event{Type: EventLinkUp, Link: link}

	links := peer.Links()
	if len(links) == 1 {
		evCh <- Event{Type: EventPeerLinked, Peer: peer}

		// set a timeout
		sig.On(ctx, sig.Idle(ctx, peer, defaultPeerIdleTimeout), func() {
			for l := range peer.Links() {
				l.Close()
			}
		})

		// if we only have an incoming link over tor, try to link back via other networks
		l := <-links
		if (l.Outbound() == false) && (l.Network() == tor.NetworkName) {
			go n.connect(ctx, peer)
		}
	}

	return nil
}

func (n *Network) linkerMutex(p *peer.Peer) *sync.Mutex {
	n.peersMu.Lock()
	defer n.peersMu.Unlock()

	if mu, ok := n.linkerMu[p]; ok {
		return mu
	}
	n.linkerMu[p] = &sync.Mutex{}
	return n.linkerMu[p]
}

func (n *Network) connect(ctx context.Context, p *peer.Peer) {
	mu := n.linkerMutex(p)
	mu.Lock()
	defer mu.Unlock()

	// prepare the contacts resolver without already linked networks
	var resolver contacts.Resolver = n.Contacts
	for _, name := range link.Networks(p.Links()) {
		resolver = contacts.Filter(resolver, contacts.SkipNetwork(name))
	}

	linker := linker.ConcurrentLinker{
		LocalID:  n.localID,
		RemoteID: p.Identity(),
		Resolver: resolver,
	}

	l := linker.Link(ctx)
	if l != nil {
		n.newLinks <- l
	}
}

func (n *Network) Connect(parent context.Context, p *peer.Peer) (*link.Link, error) {
	ctx, cancel := context.WithTimeout(parent, time.Minute)
	defer cancel()

	// see if we have a link already
	if l := link.Select(p.Links(), link.Fastest); l != nil {
		return l, nil
	}

	ch := make(chan *link.Link, 1)

	// wait for a link with the peer
	go func() {
		defer close(ch)
		for {
			select {
			case <-ctx.Done():
				return
			case <-p.StateQueue().Wait():
				if l := link.Select(p.Links(), link.Fastest); l != nil {
					ch <- l
					return
				}
			}
		}
	}()

	// try to produce a link using the default linker
	go n.connect(ctx, p)

	l, ok := <-ch
	if !ok {
		return nil, errors.New("peer unreachable")
	}

	return l, nil
}
