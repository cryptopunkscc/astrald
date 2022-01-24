package api

import (
	"io"
	"sync"
)

type Subscriptions struct {
	sync.Mutex
	Set map[io.WriteCloser]struct{}
}

func NewSubscriptions() *Subscriptions {
	return &Subscriptions{Set: map[io.WriteCloser]struct{}{}}
}

func (s *Subscriptions) Subscribe(w io.WriteCloser) Unsubscribe {
	s.Lock()
	s.Set[w] = struct{}{}
	s.Unlock()
	return func() {
		s.Lock()
		delete(s.Set, w)
		s.Unlock()
	}
}

type Unsubscribe func()
