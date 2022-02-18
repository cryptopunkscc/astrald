package node

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/hub"
	alink "github.com/cryptopunkscc/astrald/link"
	"github.com/cryptopunkscc/astrald/node/config"
	"github.com/cryptopunkscc/astrald/node/contacts"
	"github.com/cryptopunkscc/astrald/node/event"
	"github.com/cryptopunkscc/astrald/node/infra"
	nlink "github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/node/linking"
	"github.com/cryptopunkscc/astrald/node/peer"
	"github.com/cryptopunkscc/astrald/node/presence"
	"github.com/cryptopunkscc/astrald/node/server"
	"github.com/cryptopunkscc/astrald/nodeinfo"
	"github.com/cryptopunkscc/astrald/sig"
	"github.com/cryptopunkscc/astrald/storage"
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
	Linking  *linking.Manager
	Store    storage.Store
	Presence *presence.Presence
	Peers    *peer.Manager

	// peers
	peers   map[string]*peer.Peer
	peersMu sync.Mutex

	events  event.Queue
	links   chan *alink.Link
	queries chan *nlink.Query
}

func (node *Node) Identity() id.Identity {
	return node.identity
}

func (node *Node) Alias() string {
	return node.Config.GetAlias()
}

func (node *Node) Subscribe(cancel sig.Signal) <-chan event.Event {
	return node.events.Subscribe(cancel)
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
	go node.processLinks(ctx)
	go node.processQueries(ctx)

	node.events.On(ctx.Done(), peer.EventLinked{}, func(e event.Event) {
		event := e.(peer.EventLinked)

		sig.OnCtx(ctx, sig.Idle(ctx, event.Peer, defaultPeerIdleTimeout), event.Peer.Unlink)
	})

	// log all node events
	go func() {
		for event := range node.events.Subscribe(ctx.Done()) {
			node.logEvent(event)
		}
	}()

	for {
		select {
		case l := <-node.Server.Links():
			node.AddLink(l)

		case l := <-node.Linking.Links():
			node.AddLink(l)

		case <-ctx.Done():
			return
		}
	}
}
