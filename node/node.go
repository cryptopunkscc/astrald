package node

import (
	"context"
	"fmt"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/hub"
	"github.com/cryptopunkscc/astrald/infra/tor"
	"github.com/cryptopunkscc/astrald/node/config"
	"github.com/cryptopunkscc/astrald/node/contacts"
	"github.com/cryptopunkscc/astrald/node/event"
	"github.com/cryptopunkscc/astrald/node/infra"
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/node/linker"
	"github.com/cryptopunkscc/astrald/node/peer"
	"github.com/cryptopunkscc/astrald/node/presence"
	"github.com/cryptopunkscc/astrald/node/server"
	"github.com/cryptopunkscc/astrald/nodeinfo"
	"github.com/cryptopunkscc/astrald/sig"
	"github.com/cryptopunkscc/astrald/storage"
	"io"
	"log"
	"sync"
	"time"
)

const defaultQueryTimeout = time.Minute
const defaultPeerIdleTimeout = 5 * time.Minute

type Node struct {
	Config   config.Config
	identity id.Identity

	Infra    *infra.Infra
	Contacts *contacts.Manager
	Ports    *hub.Hub
	Server   *server.Server
	Linker   *linker.Manager
	Store    storage.Store
	Presence *presence.Presence

	// peers
	peers   map[string]*peer.Peer
	peersMu sync.Mutex
	queries chan *link.Query
	events  *event.Queue
}

// Run starts the node, waits for it to finish and returns an error if any
func Run(ctx context.Context, dataDir string, modules ...ModuleRunner) (*Node, error) {
	var err error

	// Storage
	store := storage.NewFilesystemStorage(dataDir)

	// Config
	cfg, err := config.Load(store)
	if err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	node := &Node{
		Config:  cfg,
		Store:   store,
		Ports:   hub.New(),
		peers:   make(map[string]*peer.Peer),
		queries: make(chan *link.Query),
		events:  event.NewQueue(),
	}

	// Set up identity
	if err := node.setupIdentity(); err != nil {
		return nil, fmt.Errorf("identity setup error: %w", err)
	}

	// Say hello
	nodeKey := node.identity.PublicKeyHex()
	if node.Alias() != "" {
		nodeKey = fmt.Sprintf("%s (%s)", node.Alias(), nodeKey)
	}
	log.Printf("astral node %s statrting...", nodeKey)

	// Set up the infrastructure
	node.Infra, err = infra.New(
		node.Config.Infra,
		infra.FilteredQuerier{Querier: node, FilteredID: node.identity},
		node.Store,
	)
	if err != nil {
		log.Println("infra error:", err)
		return nil, err
	}

	// Contacts
	node.Contacts = contacts.New(store)

	// Linker
	node.Linker = linker.New(node.identity, node.Contacts, node.Infra)

	// Server
	node.Server, err = server.Run(ctx, node.identity, node.Infra)
	if err != nil {
		return nil, err
	}

	// Modules
	node.runModules(ctx, modules)

	// Presence
	node.Presence = presence.Run(ctx, node.Infra)

	err = node.Presence.Announce(ctx, node.identity)
	if err != nil {
		log.Println("announce error:", err) // non-critical
	}

	// Run it
	go node.process(ctx)

	return node, nil
}

func (node *Node) AddLink(ctx context.Context, link *link.Link) error {
	if link == nil {
		panic("link is nil")
	}

	peer := node.makePeer(link.RemoteIdentity())

	if err := peer.Add(link); err != nil {
		return err
	}

	// forward link's requests
	go func() {
		for query := range link.Queries() {
			node.queries <- query
		}
		node.pushEvent(Event{Type: EventLinkDown, Link: link})

		if len(peer.Links()) == 0 {
			node.pushEvent(Event{Type: EventPeerUnlinked, Peer: peer})
		}
	}()

	node.pushEvent(Event{Type: EventLinkUp, Link: link})

	links := peer.Links()
	if len(links) == 1 {
		node.pushEvent(Event{Type: EventPeerLinked, Peer: peer})

		// set a timeout
		sig.On(ctx, sig.Idle(ctx, peer, defaultPeerIdleTimeout), func() {
			for l := range peer.Links() {
				l.Close()
			}
		})

		// if we only have an incoming link over tor, try to link back via other networks
		l := <-links
		if (l.Outbound() == false) && (l.Network() == tor.NetworkName) {
			go node.Linker.NewLink(ctx, peer)
		}
	}

	return nil
}

func (node *Node) Identity() id.Identity {
	return node.identity
}

func (node *Node) Alias() string {
	return node.Config.GetAlias()
}

func (node *Node) Follow(ctx context.Context) <-chan event.Eventer {
	return node.events.Follow(ctx)
}

func (node *Node) NodeInfo() *nodeinfo.NodeInfo {
	i := nodeinfo.New(node.identity)
	i.Alias = node.Alias()

	for _, a := range node.Infra.Addresses() {
		if a.Global {
			i.Addresses = append(i.Addresses, a.Addr)
		}
	}

	return i
}

func (node *Node) process(ctx context.Context) {
	for {
		select {
		case e := <-node.Presence.Events():
			node.pushEvent(e)

		case l := <-node.Server.Links():
			node.AddLink(ctx, l)

		case l := <-node.Linker.Links():
			node.AddLink(ctx, l)

		case query := <-node.queries:
			go node.serveQuery(query)

		case <-ctx.Done():
			return
		}
	}
}

func (node *Node) serveQuery(query *link.Query) {
	if query.String() == ".ping" {
		log.Printf("[%s] ping\n", node.Contacts.DisplayName(query.Caller()))
		query.Reject()
		return
	}

	log.Printf("[%s] <- %s (%s)\n", node.Contacts.DisplayName(query.Caller()), query, query.Link().Network())

	// Query a session with the service
	localStream, err := node.Ports.Query(query.String(), query.Caller())
	if err != nil {
		query.Reject()
		return
	}

	// Accept remote party's query
	remoteStream, err := query.Accept()
	if err != nil {
		localStream.Close()
		return
	}

	// Connect local and remote streams
	go func() {
		_, _ = io.Copy(localStream, remoteStream)
		_ = localStream.Close()
	}()
	go func() {
		_, _ = io.Copy(remoteStream, localStream)
		_ = remoteStream.Close()
	}()
}

func (node *Node) pushEvent(event event.Eventer) {
	switch event := event.(type) {
	case Event:
		node.handleNetworkEvent(event)

	case presence.Event:
		node.handlePresenceEvent(event)
	}

	node.events = node.events.Push(event)
}

func (node *Node) handleNetworkEvent(event Event) {
	switch event.Event() {
	case EventPeerLinked:
		log.Printf("[%s] linked\n", node.Contacts.DisplayName(event.Peer.Identity()))
	case EventPeerUnlinked:
		log.Printf("[%s] unlinked\n", node.Contacts.DisplayName(event.Peer.Identity()))
	}
}

func (node *Node) handlePresenceEvent(event presence.Event) {
	switch event.Event() {
	case presence.EventIdentityPresent:
		log.Printf("[%s] present (%s)\n", node.Contacts.DisplayName(event.Identity()), event.Addr().Network())
		node.Contacts.Find(event.Identity(), true).Add(event.Addr())
		node.Contacts.Save()

	case presence.EventIdentityGone:
		log.Printf("[%s] gone\n", node.Contacts.DisplayName(event.Identity()))
	}
}
