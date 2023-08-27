package network

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
	"sync"
	"sync/atomic"
	"time"
)

type LinkSet struct {
	mu     sync.Mutex
	links  map[int]*ActiveLink
	index  map[net.Link]int
	nextID atomic.Int64
}

type ActiveLink struct {
	id    int
	added time.Time
	net.Link
}

func (a *ActiveLink) ID() int {
	return a.id
}

func (a *ActiveLink) AddedAt() time.Time {
	return a.added
}

func NewLinkSet() *LinkSet {
	return &LinkSet{
		links: make(map[int]*ActiveLink, 0),
		index: make(map[net.Link]int),
	}
}

// Add adds a link to the set.
// Possible errors: ErrDuplicateLink
func (set *LinkSet) Add(l net.Link) (*ActiveLink, error) {
	set.mu.Lock()
	defer set.mu.Unlock()

	if l == nil {
		return nil, ErrLinkIsNil
	}

	if _, found := set.index[l]; found {
		return nil, ErrDuplicateLink
	}

	active := &ActiveLink{
		id:    int(set.nextID.Add(1)),
		added: time.Now(),
		Link:  l,
	}

	set.add(active)

	return active, nil
}

func (set *LinkSet) add(active *ActiveLink) {
	set.links[active.id] = active
	set.index[active.Link] = active.id
}

func (set *LinkSet) Find(id int) (*ActiveLink, error) {
	set.mu.Lock()
	defer set.mu.Unlock()

	return set.find(id)
}

func (set *LinkSet) find(id int) (*ActiveLink, error) {
	if active, found := set.links[id]; found {
		return active, nil
	}

	return nil, ErrLinkNotFound
}

// Remove removes a link from the set.
// Errors: ErrLinkNotFound
func (set *LinkSet) Remove(id int) error {
	set.mu.Lock()
	defer set.mu.Unlock()

	l, err := set.find(id)
	if err != nil {
		return err
	}

	delete(set.index, l.Link)
	delete(set.links, l.id)

	return nil
}

// All returns a copy of an array holding all links in the set
func (set *LinkSet) All() []*ActiveLink {
	set.mu.Lock()
	defer set.mu.Unlock()

	links := make([]*ActiveLink, 0, len(set.links))
	for _, l := range set.links {
		links = append(links, l)
	}

	return links
}

func (set *LinkSet) AllRaw() []net.Link {
	set.mu.Lock()
	defer set.mu.Unlock()

	links := make([]net.Link, len(set.links))
	for _, l := range set.links {
		links = append(links, l.Link)
	}

	return links
}

func (set *LinkSet) Count() int {
	set.mu.Lock()
	defer set.mu.Unlock()

	return len(set.links)
}

func (set *LinkSet) ByRemoteIdentity(identity id.Identity) *LinkSet {
	set.mu.Lock()
	defer set.mu.Unlock()

	var subset = NewLinkSet()

	for _, l := range set.links {
		if l.RemoteIdentity().IsEqual(identity) {
			subset.add(l)
		}
	}

	return subset
}

func (set *LinkSet) ByLocalIdentity(identity id.Identity) *LinkSet {
	set.mu.Lock()
	defer set.mu.Unlock()

	var subset = NewLinkSet()

	for _, l := range set.links {
		if l.LocalIdentity().IsEqual(identity) {
			subset.add(l)
		}
	}

	return subset
}
