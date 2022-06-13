package node

import (
	"github.com/cryptopunkscc/astrald/hub"
	"github.com/cryptopunkscc/astrald/logfmt"
	"github.com/cryptopunkscc/astrald/node/event"
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/node/peer"
	"github.com/cryptopunkscc/astrald/node/presence"
	"log"
	"reflect"
)

func (node *Node) logEvent(event event.Event) {
	var displayName = node.Contacts.DisplayName

	eventName := reflect.TypeOf(event).String()

	ok := node.Config.LogEventsInclude(eventName) || node.Config.LogEventsInclude("*")

	if !ok {
		return
	}

	switch event := event.(type) {
	case peer.EventLinked:
		log.Printf("%s linked\n", displayName(event.Peer.Identity()))

	case peer.EventUnlinked:
		log.Printf("%s unlinked\n", displayName(event.Peer.Identity()))

	case peer.EventLinkEstablished:
		log.Printf(
			"link with %s established over %s",
			displayName(event.Link.RemoteIdentity()),
			event.Link.Network(),
		)

	case peer.EventLinkClosed:
		log.Printf(
			"link with %s over %s closed",
			displayName(event.Link.RemoteIdentity()),
			event.Link.Network(),
		)

	case presence.EventIdentityPresent:
		log.Printf("%s present (%s)\n", displayName(event.Identity), event.Addr.Network())

	case presence.EventIdentityGone:
		log.Printf("%s gone\n", displayName(event.Identity))

	case link.EventConnEstablished:
		c := event.Conn
		log.Printf(
			"%s: %s%s open\n",
			displayName(c.Link().RemoteIdentity()),
			logfmt.Bool(c.Outbound(), "->", "<-"),
			c.Query(),
		)

	case link.EventConnClosed:
		c := event.Conn
		log.Printf(
			"%s: %s%s closed\n",
			displayName(c.Link().RemoteIdentity()),
			logfmt.Bool(c.Outbound(), "->", "<-"),
			c.Query(),
		)

	case hub.EventPortRegistered:
		log.Printf("port registered: %s\n", event.PortName)

	case hub.EventPortReleased:
		log.Printf("port released: %s\n", event.PortName)

	default:
		log.Println("event:", reflect.TypeOf(event).String())
	}
}
