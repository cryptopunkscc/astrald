package tcp

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"io"
)

const ModuleName = "tcp"

type Module interface {
	exonet.Dialer
	exonet.Unpacker
	exonet.Parser
	ListenPort() int
}

type EventNetworkAddressChanged struct {
	Removed []string
	Added   []string
	All     []string
}

func (EventNetworkAddressChanged) ObjectType() string {
	return "astrald.mod.tcp.events.network_address_changed"
}

func (e EventNetworkAddressChanged) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(e).WriteTo(w)
}

func (e *EventNetworkAddressChanged) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(e).ReadFrom(r)
}
