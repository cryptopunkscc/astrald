package link

import (
	"errors"
	"sync"
)

type Set struct {
	links map[*Link]struct{}
	mu    sync.Mutex
}

func NewSet() *Set {
	return &Set{
		links: make(map[*Link]struct{}, 0),
	}
}

func (set *Set) Add(link *Link) error {
	set.mu.Lock()
	defer set.mu.Unlock()

	if _, found := set.links[link]; found {
		return errors.New("duplicate item")
	}

	set.links[link] = struct{}{}

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

	delete(set.links, link)

	return nil
}

func (set *Set) Count() int {
	set.mu.Lock()
	defer set.mu.Unlock()

	return len(set.links)
}

func (set *Set) Each() <-chan *Link {
	set.mu.Lock()
	defer set.mu.Unlock()

	ch := make(chan *Link, len(set.links))
	for link, _ := range set.links {
		ch <- link
	}
	close(ch)
	return ch
}
