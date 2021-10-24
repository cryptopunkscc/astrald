package link

import (
	"errors"
	"sync"
)

type Set struct {
	list   []*Link
	listMu sync.Mutex

	watches   []chan *Link
	watchesMu sync.Mutex
}

func NewSet() *Set {
	return &Set{
		list:    make([]*Link, 0),
		watches: make([]chan *Link, 0),
	}
}

type CancelFunc func()

func (set *Set) All() <-chan *Link {
	set.listMu.Lock()
	defer set.listMu.Unlock()

	ch := make(chan *Link, len(set.list))
	for _, item := range set.list {
		ch <- item
	}
	close(ch)

	return ch
}

func (set *Set) Add(link *Link) error {
	set.listMu.Lock()
	defer set.listMu.Unlock()

	for _, l := range set.list {
		if l == link {
			return errors.New("duplicate link")
		}
	}

	set.list = append(set.list, link)

	go func() {
		<-link.WaitClose()
		set.remove(link)
	}()

	go set.notify(link)

	return nil
}

func (set *Set) Watch(includeAll bool) (<-chan *Link, CancelFunc) {
	set.watchesMu.Lock()
	defer set.watchesMu.Unlock()

	var ch chan *Link

	if includeAll {
		set.listMu.Lock()
		ch = make(chan *Link, len(set.list)+1)
		for _, item := range set.list {
			ch <- item
		}
		set.listMu.Unlock()
	} else {
		ch = make(chan *Link, 1)
	}

	set.watches = append(set.watches, ch)

	return ch, func() {
		set.cancelWatch(ch)
	}
}

func (set *Set) cancelWatch(ch chan *Link) {
	set.watchesMu.Lock()
	defer set.watchesMu.Unlock()

	for idx, w := range set.watches {
		if w == ch {
			set.watches = append(set.watches[:idx], set.watches[idx+1:]...)
			close(ch)
			return
		}
	}
}

func (set *Set) notify(lnk *Link) {
	set.watchesMu.Lock()
	defer set.watchesMu.Unlock()

	for _, ch := range set.watches {
		ch <- lnk
	}
}

func (set *Set) remove(link *Link) {
	set.listMu.Lock()
	defer set.listMu.Unlock()

	for i, l := range set.list {
		if l == link {
			set.list = append(set.list[:i], set.list[i+1:]...)
			return
		}
	}
}
