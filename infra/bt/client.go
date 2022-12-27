package bt

import (
	"context"
	"github.com/cryptopunkscc/astrald/infra"
)

var Instance Client = &Bluetooth{}

type Client interface {
	infra.Network
	Dial(context.Context, Addr) (infra.Conn, error)
}
