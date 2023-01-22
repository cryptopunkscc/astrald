package node

import (
	"context"
	"fmt"
	"github.com/cryptopunkscc/astrald/hub"
	"github.com/cryptopunkscc/astrald/logfmt"
	"github.com/cryptopunkscc/astrald/node/event"
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/node/peers"
	"github.com/cryptopunkscc/astrald/node/presence"
	"log"
	"reflect"
	"time"
)

func (node *Node) handleEvents(ctx context.Context) error {
	for event := range node.events.Subscribe(ctx) {
		node.logEvent(event)

		switch event := event.(type) {
		case presence.EventIdentityPresent:
			node.Tracker.Add(event.Identity, event.Addr, time.Now().Add(60*time.Minute))
		}
	}

	return nil
}

func (node *Node) logEvent(event event.Event) {
	var displayName = node.Contacts.DisplayName

	eventName := reflect.TypeOf(event).String()

	ok := node.Config.Log.IsEventLoggable(eventName)

	if !ok {
		return
	}

	switch event := event.(type) {
	case peers.EventPeerLinked:
		log.Printf("%s linked\n", displayName(event.Peer.Identity()))

	case peers.EventPeerUnlinked:
		log.Printf("%s unlinked\n", displayName(event.Peer.Identity()))

	case link.EventLinkEstablished:
		log.Printf(
			"link with %s established over %s",
			displayName(event.Link.RemoteIdentity()),
			event.Link.Network(),
		)

	case link.EventLinkClosed:
		log.Printf(
			"link with %s over %s closed (%s)",
			displayName(event.Link.RemoteIdentity()),
			event.Link.Network(),
			event.Link.Err(),
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
		if stringer, ok := event.(fmt.Stringer); ok {
			log.Printf("<%s> %s\n", reflect.TypeOf(event).String(), stringer.String())
		} else {
			log.Printf("<%s>\n", reflect.TypeOf(event).String())
		}
	}
}
