package infra

import (
	"errors"
	"github.com/cryptopunkscc/astrald/net"
)

// Endpoints returns a copy of a list of all infrastructural endpoints
func (infra *CoreInfra) Endpoints() []net.Endpoint {
	infra.mu.Lock()
	defer infra.mu.Unlock()

	var endpoints = make([]net.Endpoint, 0)

	// collect addresses from all enabled drivers
	for name, drv := range infra.networkDrivers {
		if !infra.config.driversContain(name) {
			continue
		}

		if lister, ok := drv.(EndpointLister); ok {
			endpoints = append(endpoints, lister.Endpoints()...)
		}
	}

	for _, e := range infra.endpoints {
		endpoints = append(endpoints, e.Endpoints()...)
	}

	return endpoints
}

func (infra *CoreInfra) AddEndpoints(e EndpointLister) error {
	infra.mu.Lock()
	defer infra.mu.Unlock()

	for _, i := range infra.endpoints {
		if i == e {
			return errors.New("already added")
		}
	}

	infra.endpoints = append(infra.endpoints, e)
	return nil
}

func (infra *CoreInfra) RemoveEndpoints(e EndpointLister) error {
	infra.mu.Lock()
	defer infra.mu.Unlock()

	for i, v := range infra.endpoints {
		if v == e {
			infra.endpoints = append(infra.endpoints[:i], infra.endpoints[i+1:]...)
			return nil
		}
	}

	return errors.New("not found")
}
