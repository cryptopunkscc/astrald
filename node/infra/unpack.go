package infra

import (
	"errors"
	"github.com/cryptopunkscc/astrald/net"
)

func (infra *CoreInfra) Unpack(network string, data []byte) (net.Endpoint, error) {
	if n, found := infra.networkDrivers[network]; found {
		if unpacker, ok := n.(Unpacker); ok {
			return unpacker.Unpack(network, data)
		}
	}

	if unpacker, found := infra.unpackers[network]; found {
		return unpacker.Unpack(network, data)
	}

	return net.NewGenericEndpoint(network, data), nil
}

func (infra *CoreInfra) Parse(network string, address string) (net.Endpoint, error) {
	if n, found := infra.networkDrivers[network]; found {
		if unpacker, ok := n.(Parser); ok {
			return unpacker.Parse(network, address)
		}
	}

	if n, found := infra.unpackers[network]; found {
		if unpacker, ok := n.(Parser); ok {
			return unpacker.Parse(network, address)
		}
	}

	return nil, errors.New("unsupported network")
}

func (infra *CoreInfra) AddUnpacker(network string, unpacker Unpacker) error {
	infra.mu.Lock()
	defer infra.mu.Unlock()

	if _, found := infra.unpackers[network]; found {
		return errors.New("already added")
	}

	infra.unpackers[network] = unpacker

	return nil
}

func (infra *CoreInfra) RemoveUnpacker(network string) error {
	infra.mu.Lock()
	defer infra.mu.Unlock()

	if _, found := infra.unpackers[network]; !found {
		return errors.New("not found")
	}

	delete(infra.unpackers, network)

	return nil
}
