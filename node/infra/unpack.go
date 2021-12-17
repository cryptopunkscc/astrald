package infra

import "github.com/cryptopunkscc/astrald/infra"

func (i *Infra) Unpack(network string, data []byte) (infra.Addr, error) {
	for net := range i.Networks() {
		if unpacker, ok := net.(infra.Unpacker); ok {
			if addr, err := unpacker.Unpack(network, data); err == nil {
				return addr, nil
			}
		}
	}

	return NewUnsupportedAddr(network, data), nil
}
