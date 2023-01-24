package node

import (
	"context"
	"fmt"
	"github.com/cryptopunkscc/astrald/node/event"
	"github.com/cryptopunkscc/astrald/node/presence"
	"reflect"
	"time"
)

func (node *Node) handleEvents(ctx context.Context) error {
	for event := range node.Events.Subscribe(ctx) {
		node.logEvent(event)

		switch event := event.(type) {
		case presence.EventIdentityPresent:
			node.Tracker.Add(event.Identity, event.Addr, time.Now().Add(60*time.Minute))
		}
	}

	return nil
}

func (node *Node) logEvent(ev event.Event) {
	var eventName = reflect.TypeOf(ev).String()

	if !node.Config.Log.IsEventLoggable(eventName) {
		return
	}

	if stringer, ok := ev.(fmt.Stringer); ok {
		log.Log("%s<%s>%s %s%s", log.Purple(), reflect.TypeOf(ev).String(), log.Gray(), stringer.String(), log.Reset())
	} else {
		log.Log("%s<%s>%s", log.Purple(), reflect.TypeOf(ev).String(), log.Reset())
	}
}
