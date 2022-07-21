package bt

import (
	"github.com/cryptopunkscc/astrald/infra"
)

const NetworkName = "bt"

var Instance Client = Bluetooth{}

type Client interface {
	infra.Network
	infra.AddrLister
	infra.Unpacker
	infra.Listener
	infra.Dialer
}
