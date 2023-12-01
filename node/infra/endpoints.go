package infra

import (
	"errors"
	"github.com/cryptopunkscc/astrald/net"
)

// Endpoints returns a copy of a list of all infrastructural endpoints
func (infra *CoreInfra) Endpoints() []net.Endpoint {
	infra.mu.RLock()
	defer infra.mu.RUnlock()

	var endpoints = make([]net.Endpoint, 0)

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
