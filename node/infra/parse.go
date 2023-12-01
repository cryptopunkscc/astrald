package infra

import (
	"github.com/cryptopunkscc/astrald/net"
)

func (infra *CoreInfra) Parse(network string, address string) (net.Endpoint, error) {
	infra.mu.RLock()
	defer infra.mu.RUnlock()

	if parser, found := infra.parsers[network]; found {
		return parser.Parse(network, address)
	}

	if n, found := infra.unpackers[network]; found {
		if unpacker, ok := n.(Parser); ok {
			return unpacker.Parse(network, address)
		}
	}

	return nil, ErrUnsupportedNetwork
}

func (infra *CoreInfra) SetParser(network string, parser Parser) {
	infra.mu.Lock()
	defer infra.mu.Unlock()

	if parser == nil {
		delete(infra.parsers, network)
	} else {
		infra.parsers[network] = parser
	}
}
