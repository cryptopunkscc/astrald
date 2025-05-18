package tcp

import (
	"github.com/cryptopunkscc/astrald/astral"
	"io"
)

type EventNetworkAddressChanged struct {
	Removed []string
	Added   []string
	All     []string
}

// astral

func (EventNetworkAddressChanged) ObjectType() string {
	return "mod.tcp.events.network_address_changed"
}

func (e EventNetworkAddressChanged) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(e).WriteTo(w)
}

func (e *EventNetworkAddressChanged) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(e).ReadFrom(r)
}

// ...

func init() {
	_ = astral.DefaultBlueprints.Add(&EventNetworkAddressChanged{})
}
