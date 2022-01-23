package warpdrive

import (
	"io"
	"sync"
)

type subscriptions struct {
	sync.Mutex
	set map[io.WriteCloser]struct{}
}

func newSubscriptions() *subscriptions {
	return &subscriptions{set: map[io.WriteCloser]struct{}{}}
}

func (s *subscriptions) subscribe(w io.WriteCloser) unsubscribe {
	s.Lock()
	s.set[w] = struct{}{}
	s.Unlock()
	return func() {
		s.Lock()
		delete(s.set, w)
		s.Unlock()
	}
}

type unsubscribe func()
