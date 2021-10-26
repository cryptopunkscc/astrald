package link

import (
	"errors"
	"sync"
)

type Set struct {
	links []*Link
	mu    sync.Mutex
}

func NewSet() *Set {
	return &Set{
		links: make([]*Link, 0),
	}
}

func (set *Set) Add(link *Link) error {
	set.mu.Lock()
	defer set.mu.Unlock()

	if set.index(link) != -1 {
		return errors.New("duplicate item")
	}

	set.links = append(set.links, link)

	return nil
}

func (set *Set) Index(link *Link) int {
	set.mu.Lock()
	defer set.mu.Unlock()

	return set.index(link)
}

func (set *Set) index(link *Link) int {
	for i, l := range set.links {
		if l == link {
			return i
		}
	}
	return -1
}

func (set *Set) Contains(link *Link) bool {
	set.mu.Lock()
	defer set.mu.Unlock()

	return set.index(link) >= 0
}

func (set *Set) Remove(link *Link) error {
	set.mu.Lock()
	defer set.mu.Unlock()

	i := set.index(link)
	if i == -1 {
		return errors.New("not found")
	}

	set.links = append(set.links[:i], set.links[i+1:]...)

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
	for _, link := range set.links {
		ch <- link
	}
	close(ch)
	return ch
}
