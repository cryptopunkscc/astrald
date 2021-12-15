package contacts

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

func (f *FilteredResolver) Lookup(nodeID id.Identity) <-chan infra.Addr {
	var addrs = f.parent.Lookup(nodeID)

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

func SkipNetwork(network string) FilterFunc {
	return func(addr infra.Addr) bool {
		return addr.Network() != network
	}
}
