package bt

import (
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/infra/drivers/bt/bluez"
)

func (*Driver) Endpoints() []net.Endpoint {
	endpoints := make([]net.Endpoint, 0)

	b, err := bluez.New()
	if err != nil {
		return endpoints
	}

	adapters, err := b.Adapters()
	if err != nil {
		return endpoints
	}

	for _, adapter := range adapters {
		if a, err := adapter.Address(); err == nil {
			if parsed, err := Parse(a); err == nil {
				endpoints = append(endpoints, parsed)
			}
		}
	}

	return endpoints
}
