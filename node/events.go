package node

import (
	"context"
	"fmt"
	"github.com/cryptopunkscc/astrald/node/event"
	"reflect"
)

func (node *CoreNode) handleEvents(ctx context.Context) error {
	for e := range node.events.Subscribe(ctx) {
		node.logEvent(e)
	}

	return nil
}

func (node *CoreNode) logEvent(e event.Event) {
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
