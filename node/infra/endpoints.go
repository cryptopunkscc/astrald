package infra

import "github.com/cryptopunkscc/astrald/net"

// Endpoints returns a copy of a list of all infrastructural endpoints
func (i *CoreInfra) Endpoints() []net.Endpoint {
	var endpoints = make([]net.Endpoint, 0)

	// collect addresses from all enabled drivers
	for name, drv := range i.networkDrivers {
		if !i.config.driversContain(name) {
			continue
		}

		if lister, ok := drv.(EndpointLister); ok {
			endpoints = append(endpoints, lister.Endpoints()...)
		}
	}

	return endpoints
}
