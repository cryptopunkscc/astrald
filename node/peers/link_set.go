package peers

import (
	"github.com/cryptopunkscc/astrald/node/link"
	"sync"
)

type LinkSet struct {
	mu    sync.Mutex
	links []*link.Link
	index map[*link.Link]struct{}
}

func NewLinkSet() *LinkSet {
	return &LinkSet{
		links: make([]*link.Link, 0),
		index: make(map[*link.Link]struct{}),
	}
}

// Add adds a link to the set.
// Possible errors: ErrDuplicateLink
func (set *LinkSet) Add(l *link.Link) error {
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
func (set *LinkSet) Remove(l *link.Link) error {
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
func (set *LinkSet) Contains(l *link.Link) (found bool) {
	set.mu.Lock()
	defer set.mu.Unlock()

	_, found = set.index[l]
	return
}

// All returns a copy of an array holding all links in the set
func (set *LinkSet) All() []*link.Link {
	set.mu.Lock()
	defer set.mu.Unlock()

	links := make([]*link.Link, len(set.links))
	copy(links, set.links)

	return links
}

func (set *LinkSet) Count() int {
	return len(set.links)
}
