package core

import (
	"context"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/events"
	"github.com/cryptopunkscc/astrald/node"
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
		node.log.Log("<%s> %s", reflect.TypeOf(e).String(), stringer.String())
	} else {
		node.log.Log("<%s>", reflect.TypeOf(e).String())
	}
}

type EventLinkAdded struct {
	Link node.ActiveLink
}

type EventLinkRemoved struct {
	Link node.ActiveLink
}

type EventConnAdded struct {
	Conn *MonitoredConn
}

type EventConnRemoved struct {
	Conn *MonitoredConn
}

var ErrUnsupportedNetwork = errors.New("unsupported network")
