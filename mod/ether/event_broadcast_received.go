package ether

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/tcp"
	"github.com/cryptopunkscc/astrald/streams"
	"io"
)

var _ astral.Object = &EventBroadcastReceived{}

type EventBroadcastReceived struct {
	SourceID *astral.Identity
	SourceIP tcp.IP
	Object   astral.Object
}

func (EventBroadcastReceived) ObjectType() string {
	return "astrald.mod.ether.event_broadcast_received"
}

func (e EventBroadcastReceived) WriteTo(w io.Writer) (n int64, err error) {
	return streams.WriteAllTo(w, e.SourceID, e.SourceIP, e.Object)
}

func (e *EventBroadcastReceived) ReadFrom(r io.Reader) (n int64, err error) {
	e.SourceID = &astral.Identity{}
	return streams.ReadAllFrom(r, e.SourceID, &e.SourceIP, e.Object)
}
