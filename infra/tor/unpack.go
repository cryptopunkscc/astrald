package tor

import "github.com/cryptopunkscc/astrald/infra"

var _ infra.Unpacker = &Tor{}

// Unpack deserializes an address from its binary format
func (tor *Tor) Unpack(network string, bytes []byte) (infra.Addr, error) {
	if network != NetworkName {
		return nil, infra.ErrUnsupportedAddress
	}

	return Unpack(bytes)
}
