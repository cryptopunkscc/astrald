package node

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/hub"
	alink "github.com/cryptopunkscc/astrald/link"
	"github.com/cryptopunkscc/astrald/node/config"
	"github.com/cryptopunkscc/astrald/node/contacts"
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

	events  *sig.Queue
	links   chan *alink.Link
	queries chan *nlink.Query
}

func (node *Node) Identity() id.Identity {
	return node.identity
}

func (node *Node) Alias() string {
	return node.Config.GetAlias()
}

func (node *Node) Follow(ctx context.Context) <-chan interface{} {
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
	go node.processLinks(ctx)
	go node.processQueries(ctx)

	for {
		select {
		case e := <-node.Presence.Events():
			node.emitEvent(e)

		case l := <-node.Server.Links():
			node.AddLink(l)

		case l := <-node.Linking.Links():
			node.AddLink(l)

		case <-ctx.Done():
			return
		}
	}
}
