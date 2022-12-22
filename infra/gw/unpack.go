package gw

import "github.com/cryptopunkscc/astrald/infra"

var _ infra.Unpacker = &Gateway{}

func (*Gateway) Unpack(network string, bytes []byte) (infra.Addr, error) {
	if network != NetworkName {
		return nil, infra.ErrUnsupportedAddress
	}

	return Unpack(bytes)
}
