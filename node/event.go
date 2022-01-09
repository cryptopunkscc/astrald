package node

import (
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/node/peer"
	"github.com/cryptopunkscc/astrald/node/presence"
	"log"
)

type Event interface{}

type EventLinkUp struct {
	Link *link.Link
}

type EventLinkDown struct {
	Link *link.Link
}

type EventPeerLinked struct {
	Peer *peer.Peer
	Link *link.Link
}

type EventPeerUnlinked struct {
	Peer *peer.Peer
}

func (node *Node) emitEvent(event Event) {
	node.logEvent(event)
	node.events = node.events.Push(event)
}

func (node *Node) logEvent(event Event) {
	switch event := event.(type) {
	case EventPeerLinked:
		log.Printf("[%s] linked\n", node.Contacts.DisplayName(event.Peer.Identity()))

	case EventPeerUnlinked:
		log.Printf("[%s] unlinked\n", node.Contacts.DisplayName(event.Peer.Identity()))

	case presence.EventIdentityPresent:
		log.Printf("[%s] present (%s)\n", node.Contacts.DisplayName(event.Identity), event.Addr.Network())

	case presence.EventIdentityGone:
		log.Printf("[%s] gone\n", node.Contacts.DisplayName(event.Identity))

	}
}
