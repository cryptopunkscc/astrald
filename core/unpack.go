package core

import (
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node"
)

func (infra *CoreInfra) Unpack(network string, data []byte) (net.Endpoint, error) {
	infra.mu.RLock()
	defer infra.mu.RUnlock()

	if unpacker, found := infra.unpackers[network]; found {
		return unpacker.Unpack(network, data)
	}

	return net.NewGenericEndpoint(network, data), nil
}

func (infra *CoreInfra) SetUnpacker(network string, unpacker node.Unpacker) {
	infra.mu.Lock()
	defer infra.mu.Unlock()

	if unpacker == nil {
		delete(infra.unpackers, network)
	} else {
		infra.unpackers[network] = unpacker
	}
}
