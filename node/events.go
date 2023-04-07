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
	for e := range node.events.Subscribe(ctx) {
		node.logEvent(e)

		switch e := e.(type) {
		case presence.EventIdentityPresent:
			_ = node.Tracker.Add(e.Identity, e.Addr, time.Now().Add(60*time.Minute))
		}
	}

	return nil
}

func (node *Node) logEvent(e event.Event) {
	var eventName = reflect.TypeOf(e).String()

	if !node.logConfig.IsEventLoggable(eventName) {
		return
	}

	if stringer, ok := e.(fmt.Stringer); ok {
		log.Log("%s<%s>%s %s%s", log.Purple(), reflect.TypeOf(e).String(), log.Gray(), stringer.String(), log.Reset())
	} else {
		log.Log("%s<%s>%s", log.Purple(), reflect.TypeOf(e).String(), log.Reset())
	}
}
