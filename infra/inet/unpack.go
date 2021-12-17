package inet

import "github.com/cryptopunkscc/astrald/infra"

var _ infra.Unpacker = &Inet{}

func (inet Inet) Unpack(network string, bytes []byte) (infra.Addr, error) {
	if network != NetworkName {
		return nil, infra.ErrUnsupportedAddress
	}

	return Unpack(bytes)
}
