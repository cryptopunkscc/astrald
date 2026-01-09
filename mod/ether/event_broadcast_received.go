package ether

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/ip"
)

var _ astral.Object = &EventBroadcastReceived{}

type EventBroadcastReceived struct {
	SourceID *astral.Identity
	SourceIP ip.IP
	Object   astral.Object
}

// astral

func (EventBroadcastReceived) ObjectType() string {
	return "mod.ether.events.broadcast_received"
}

func (e EventBroadcastReceived) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(e).WriteTo(w)
}

func (e *EventBroadcastReceived) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(e).ReadFrom(r)
}

func init() {
	_ = astral.Add(&EventBroadcastReceived{})
}
