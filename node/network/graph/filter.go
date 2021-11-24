package graph

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
)

var _ Resolver = &FilteredResolver{}

type FilterFunc func(addr infra.Addr) bool

type FilteredResolver struct {
	parent Resolver
	filter FilterFunc
}

func (f *FilteredResolver) Resolve(nodeID id.Identity) <-chan infra.Addr {
	var addrs = f.parent.Resolve(nodeID)

	var ch = make(chan infra.Addr, len(addrs))
	for addr := range addrs {
		if f.filter(addr) {
			ch <- addr
		}
	}
	close(ch)

	return ch
}

func Filter(parent Resolver, filter FilterFunc) *FilteredResolver {
	return &FilteredResolver{parent: parent, filter: filter}
}

func FilterNetwork(parent Resolver, network string) Resolver {
	return Filter(parent, func(addr infra.Addr) bool {
		return addr.Network() == network
	})
}
