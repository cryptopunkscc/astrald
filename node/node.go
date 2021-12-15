package node

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/hub"
	"github.com/cryptopunkscc/astrald/infra"
	iastral "github.com/cryptopunkscc/astrald/infra/astral"
	"github.com/cryptopunkscc/astrald/infra/inet"
	"github.com/cryptopunkscc/astrald/infra/tor"
	"github.com/cryptopunkscc/astrald/node/contacts"
	"github.com/cryptopunkscc/astrald/node/event"
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/node/linker"
	"github.com/cryptopunkscc/astrald/node/peer"
	"github.com/cryptopunkscc/astrald/node/presence"
	"github.com/cryptopunkscc/astrald/node/server"
	"github.com/cryptopunkscc/astrald/node/storage"
	"github.com/cryptopunkscc/astrald/nodeinfo"
	"github.com/cryptopunkscc/astrald/sig"
	"io"
	"log"
	"sync"
	"time"
)

const defaultQueryTimeout = time.Minute
const defaultPeerIdleTimeout = 5 * time.Minute

type Node struct {
	Config *Config

	Contacts *contacts.Manager
	Ports    *hub.Hub
	Server   *server.Server
	Linker   *linker.Manager
	Store    storage.Store
	Presence *presence.Presence

	identity id.Identity
	dataDir  string

	// peers
	peers   map[string]*peer.Peer
	peersMu sync.Mutex
	queries chan *link.Query
	events  *event.Queue

	// networks
	inet   *inet.Inet
	tor    *tor.Tor
	astral *iastral.Astral
}

// Run starts the node, waits for it to finish and returns an error if any
func Run(ctx context.Context, dataDir string, modules ...ModuleRunner) (*Node, error) {
	var err error

	fs := storage.NewFilesystemStorage(dataDir)

	identity := setupIdentity(fs)

	log.Printf("astral node %s statrting...", identity)

	config := loadConfig(fs)

	node := &Node{
		identity: identity,
		dataDir:  dataDir,
		Store:    fs,
		Config:   config,
		Ports:    hub.New(),
		peers:    make(map[string]*peer.Peer),
		queries:  make(chan *link.Query),
		events:   event.NewQueue(),
	}

	node.addNetworks()

	node.Contacts = contacts.New(fs)

	// start the network
	node.Linker = linker.New(node.identity, node.Contacts)

	node.Server, err = server.Run(ctx, node.identity)
	if err != nil {
		panic(err)
	}

	node.runModules(ctx, modules)

	node.Presence = presence.Run(ctx)

	err = node.Presence.Announce(ctx, node.identity)
	if err != nil {
		log.Println("announce error:", err)
	}

	go node.process(ctx)

	return node, nil
}

func (node *Node) Identity() id.Identity {
	return node.identity
}

func (node *Node) Alias() string {
	return node.Config.Alias
}

func (node *Node) Follow(ctx context.Context) <-chan event.Eventer {
	return node.events.Follow(ctx)
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

func (node *Node) NodeInfo() *nodeinfo.NodeInfo {
	i := nodeinfo.New(node.identity)
	i.Alias = node.Alias()

	for _, a := range astral.Addresses() {
		if a.Global {
			i.Addresses = append(i.Addresses, a.Addr)
		}
	}

	return i
}

func (node *Node) Info(onlyPublic bool) *contacts.Contact {
	addrs := make([]infra.Addr, 0)

	for _, a := range astral.Addresses() {
		if onlyPublic && !a.Global {
			continue
		}
		addrs = append(addrs, a.Addr)
	}

	c := contacts.NewContact(node.identity)
	for _, a := range addrs {
		c.Add(a)
	}

	if !onlyPublic {
		c.SetAlias(node.Alias())
	}

	return c
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
