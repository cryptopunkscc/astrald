package contacts

import (
	"github.com/cryptopunkscc/astrald/auth/id"
)

var _ Resolver = &FilteredResolver{}

type FilterFunc func(nodeID id.Identity, addr *Addr) bool

type FilteredResolver struct {
	parent Resolver
	filter FilterFunc
}

func (f *FilteredResolver) Lookup(nodeID id.Identity) (<-chan *Addr, error) {
	addrs, err := f.parent.Lookup(nodeID)
	if err != nil {
		return nil, err
	}

	var ch = make(chan *Addr, len(addrs))
	for addr := range addrs {
		if f.filter(nodeID, addr) {
			ch <- addr
		}
	}
	close(ch)

	return ch, nil
}

func Filter(parent Resolver, filter FilterFunc) *FilteredResolver {
	return &FilteredResolver{parent: parent, filter: filter}
}
