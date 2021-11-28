package link

import (
	"errors"
	"github.com/cryptopunkscc/astrald/sig"
	"sync"
)

type Set struct {
	*sig.Queue
	links map[*Link]struct{}
	mu    sync.Mutex
}

func NewSet() *Set {
	return &Set{
		links: make(map[*Link]struct{}, 0),
		Queue: &sig.Queue{},
	}
}

func (set *Set) Add(link *Link) error {
	set.mu.Lock()
	defer set.mu.Unlock()

	if _, found := set.links[link]; found {
		return errors.New("duplicate item")
	}

	set.links[link] = struct{}{}
	set.Queue = set.Push(link)

	return nil
}

func (set *Set) Contains(link *Link) bool {
	set.mu.Lock()
	defer set.mu.Unlock()

	_, found := set.links[link]
	return found
}

func (set *Set) Remove(link *Link) error {
	set.mu.Lock()
	defer set.mu.Unlock()

	if _, found := set.links[link]; !found {
		return errors.New("not found")
	}

	defer set.Push(link)

	delete(set.links, link)

	return nil
}

func (set *Set) Links() <-chan *Link {
	set.mu.Lock()
	defer set.mu.Unlock()

	ch := make(chan *Link, len(set.links))
	for link, _ := range set.links {
		ch <- link
	}
	close(ch)
	return ch
}
