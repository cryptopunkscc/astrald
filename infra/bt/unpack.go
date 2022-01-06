package bt

import "github.com/cryptopunkscc/astrald/infra"

var _ infra.Unpacker = &Bluetooth{}

func (bt Bluetooth) Unpack(network string, data []byte) (infra.Addr, error) {
	if network != NetworkName {
		return nil, infra.ErrUnsupportedAddress
	}

	return Unpack(data)
}
