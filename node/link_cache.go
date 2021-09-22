package node

import (
	"errors"
	"github.com/cryptopunkscc/astrald/node/link"
	"sync"
)

type LinkCache struct {
	linksMu sync.Mutex
	links   LinkPool
}

func NewLinkCache() *LinkCache {
	return &LinkCache{
		links: make(LinkPool, 0),
	}
}

func (cache *LinkCache) Links() LinkPool {
	return cache.links
}

func (cache *LinkCache) AddLink(link *link.Link) error {
	cache.linksMu.Lock()
	defer cache.linksMu.Unlock()

	for _, l := range cache.links {
		if l == link {
			return errors.New("duplicate link")
		}
	}

	cache.links = append(cache.links, link)

	go func() {
		<-link.WaitClose()

		cache.linksMu.Lock()
		defer cache.linksMu.Unlock()

		for i, l := range cache.links {
			if l == link {
				cache.links = append(cache.links[:i], cache.links[i+1:]...)
				return
			}
		}
	}()

	return nil
}
