package network

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
	"sync"
)

type LinkSet struct {
	mu    sync.Mutex
	links []net.Link
	index map[net.Link]struct{}
}

func NewLinkSet() *LinkSet {
	return &LinkSet{
		links: make([]net.Link, 0),
		index: make(map[net.Link]struct{}),
	}
}

// Add adds a link to the set.
// Possible errors: ErrDuplicateLink
func (set *LinkSet) Add(l net.Link) error {
	set.mu.Lock()
	defer set.mu.Unlock()

	if _, found := set.index[l]; found {
		return ErrDuplicateLink
	}

	set.links = append(set.links, l)
	set.index[l] = struct{}{}

	return nil
}

// Remove removes a link from the set.
// Errors: ErrLinkNotFound
func (set *LinkSet) Remove(l net.Link) error {
	set.mu.Lock()
	defer set.mu.Unlock()

	_, found := set.index[l]
	if !found {
		return ErrLinkNotFound
	}

	// lookup the index
	var idx = -1
	for i, s := range set.links {
		if l == s {
			idx = i
			break
		}
	}

	set.links = append(set.links[0:idx], set.links[idx+1:]...)
	delete(set.index, l)

	return nil
}

// Contains returns true if a link is a part of the set
func (set *LinkSet) Contains(l net.Link) (found bool) {
	set.mu.Lock()
	defer set.mu.Unlock()

	_, found = set.index[l]
	return
}

// All returns a copy of an array holding all links in the set
func (set *LinkSet) All() []net.Link {
	set.mu.Lock()
	defer set.mu.Unlock()

	links := make([]net.Link, len(set.links))
	copy(links, set.links)

	return links
}

func (set *LinkSet) Count() int {
	return len(set.links)
}

func (set *LinkSet) ByRemoteIdentity(identity id.Identity) *LinkSet {
	var subset = NewLinkSet()

	for _, l := range set.All() {
		if l.RemoteIdentity().IsEqual(identity) {
			subset.Add(l)
		}
	}

	return subset
}

func (set *LinkSet) ByLocalIdentity(identity id.Identity) *LinkSet {
	var subset = NewLinkSet()

	for _, l := range set.All() {
		if l.LocalIdentity().IsEqual(identity) {
			subset.Add(l)
		}
	}

	return subset
}
