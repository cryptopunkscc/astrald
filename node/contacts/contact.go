package contacts

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/logfmt"
	"github.com/cryptopunkscc/astrald/sig"
	"sync"
	"time"
)

const defaultAddressValidity = time.Hour * 24 * 30

type Contact struct {
	identity  id.Identity
	alias     string
	mu        sync.Mutex
	addresses []*Addr
	queue     *sig.Queue
}

func NewContact(identity id.Identity) *Contact {
	return &Contact{
		identity:  identity,
		addresses: make([]*Addr, 0),
		queue:     &sig.Queue{},
	}
}

func (c *Contact) Identity() id.Identity {
	return c.identity
}

func (c *Contact) Alias() string {
	return c.alias
}

func (c *Contact) SetAlias(alias string) {
	c.alias = alias
}

func (c *Contact) DisplayName() string {
	if c.alias != "" {
		return c.alias
	}

	return logfmt.ID(c.identity)
}

// Addr returns a channel with all addresses
func (c *Contact) Addr(follow context.Context) <-chan infra.Addr {
	c.mu.Lock()
	defer c.mu.Unlock()

	var ch chan infra.Addr

	// populate channel with addresses
	ch = make(chan infra.Addr, len(c.addresses))
	for _, addr := range c.addresses {
		ch <- addr.Addr
	}

	if follow == nil {
		close(ch)
		return ch
	}

	go func() {
		for addr := range c.queue.Subscribe(follow.Done()) {
			ch <- addr.(infra.Addr)
		}
		close(ch)
	}()

	return ch
}

func (c *Contact) Remove(addr infra.Addr) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i, a := range c.addresses {
		if infra.AddrEqual(a, addr) {
			c.addresses = append(c.addresses[:i], c.addresses[i+1:]...)
			return
		}
	}
}
