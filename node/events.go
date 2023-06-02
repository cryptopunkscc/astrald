package node

import (
	"context"
	"fmt"
	"github.com/cryptopunkscc/astrald/node/events"
	"reflect"
)

func (node *CoreNode) handleEvents(ctx context.Context) error {
	for e := range node.events.Subscribe(ctx) {
		node.logEvent(e)
	}

	return nil
}

func (node *CoreNode) logEvent(e events.Event) {
	var eventName = reflect.TypeOf(e).String()

	if !node.logConfig.IsEventLoggable(eventName) {
		return
	}

	if stringer, ok := e.(fmt.Stringer); ok {
		node.log.Log("%s<%s>%s %s%s", node.log.Purple(), reflect.TypeOf(e).String(), node.log.Gray(), stringer.String(), node.log.Reset())
	} else {
		node.log.Log("%s<%s>%s", node.log.Purple(), reflect.TypeOf(e).String(), node.log.Reset())
	}
}
