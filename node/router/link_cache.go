package router

import (
	"errors"
	"github.com/cryptopunkscc/astrald/node/auth/id"
	"github.com/cryptopunkscc/astrald/node/link"
	"sync"
)

type LinkCache struct {
	links map[id.Identity]*link.Link
	mu    sync.Mutex
}

func NewLinkCache() *LinkCache {
	return &LinkCache{
		links: make(map[id.Identity]*link.Link),
	}
}

func (pool *LinkCache) Add(l *link.Link) error {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	if pool.links == nil {
		pool.links = make(map[id.Identity]*link.Link)
	}

	if _, found := pool.links[l.RemoteIdentity()]; found {
		return errors.New("already exists")
	}

	pool.links[l.RemoteIdentity()] = l

	return nil
}

func (pool *LinkCache) Fetch(id id.Identity) *link.Link {
	return pool.links[id]
}

func (pool *LinkCache) Remove(id id.Identity) error {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	if _, found := pool.links[id]; !found {
		return errors.New("not found")
	}
	delete(pool.links, id)
	return nil
}
