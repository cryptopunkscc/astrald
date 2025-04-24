package gateway

import (
	"bytes"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mod/exonet"
)

func (mod *Module) Unpack(network string, data []byte) (exonet.Endpoint, error) {
	if network != NetworkName {
		return nil, exonet.ErrUnsupportedNetwork
	}
	return Unpack(data)
}

// Unpack converts a binary representation of the address to a struct
func Unpack(data []byte) (addr *Endpoint, err error) {
	addr = &Endpoint{TargetID: &astral.Identity{}, GatewayID: &astral.Identity{}}

	return addr, cslq.Decode(bytes.NewReader(data), "vv", addr.GatewayID, addr.TargetID)
}
