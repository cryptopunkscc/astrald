package route

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
)

var _ Router = &Filter{}

type FilterFunc func(addr infra.Addr) bool

type Filter struct {
	router Router
	filter FilterFunc
}

func NewFilter(parent Router, filter FilterFunc) *Filter {
	return &Filter{router: parent, filter: filter}
}

func (f *Filter) Route(nodeID id.Identity) *Route {
	var route = f.router.Route(nodeID)
	if route == nil {
		return nil
	}

	var filtered = New(route.Identity)
	for _, addr := range route.Addresses {
		if f.filter(addr) {
			filtered.Add(addr)
		}
	}

	return filtered
}

func FilterNetwork(parent Router, network string) Router {
	return NewFilter(parent, func(addr infra.Addr) bool {
		return addr.Network() == network
	})
}
