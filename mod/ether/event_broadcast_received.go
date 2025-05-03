package ether

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/tcp"
	"io"
)

var _ astral.Object = &EventBroadcastReceived{}

type EventBroadcastReceived struct {
	SourceID *astral.Identity
	SourceIP tcp.IP
	Object   astral.Object
}

func (EventBroadcastReceived) ObjectType() string {
	return "astrald.mod.ether.events.broadcast_received"
}

func (e EventBroadcastReceived) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(e).WriteTo(w)
}

func (e *EventBroadcastReceived) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(e).ReadFrom(r)
}

func init() {
	astral.DefaultBlueprints.Add(&EventBroadcastReceived{})
}
