package infra

import (
	"errors"
	"github.com/cryptopunkscc/astrald/net"
)

func (i *CoreInfra) Unpack(network string, data []byte) (net.Endpoint, error) {
	if n, found := i.networkDrivers[network]; found {
		if unpacker, ok := n.(Unpacker); ok {
			return unpacker.Unpack(network, data)
		}
	}

	return net.NewGenericEndpoint(network, data), nil
}

func (i *CoreInfra) Parse(network string, address string) (net.Endpoint, error) {
	if n, found := i.networkDrivers[network]; found {
		if unpacker, ok := n.(Parser); ok {
			return unpacker.Parse(network, address)
		}
	}

	return nil, errors.New("unsupported network")
}
