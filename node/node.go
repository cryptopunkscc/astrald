package node

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/hub"
	"github.com/cryptopunkscc/astrald/node/config"
	"github.com/cryptopunkscc/astrald/node/contacts"
	"github.com/cryptopunkscc/astrald/node/event"
	"github.com/cryptopunkscc/astrald/node/infra"
	"github.com/cryptopunkscc/astrald/node/peers"
	"github.com/cryptopunkscc/astrald/node/presence"
	"github.com/cryptopunkscc/astrald/nodeinfo"
	"github.com/cryptopunkscc/astrald/sig"
	"github.com/cryptopunkscc/astrald/storage"
	"time"
)

const defaultQueryTimeout = 15 * time.Second
const defaultPeerIdleTimeout = 5 * time.Minute

type Node struct {
	Config   config.Config
	identity id.Identity

	Infra    *infra.Infra
	Contacts *contacts.Manager
	Ports    *hub.Hub

	Peers    *peers.Manager
	Store    storage.Store
	Presence *presence.Presence

	events event.Queue
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
