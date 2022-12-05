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
	Addresses []*Addr
	queue     *sig.Queue
}

func NewContact(identity id.Identity) *Contact {
	return &Contact{
		identity:  identity,
		Addresses: make([]*Addr, 0),
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
	ch = make(chan infra.Addr, len(c.Addresses))
	for _, addr := range c.Addresses {
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

func (c *Contact) Add(addr infra.Addr) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, a := range c.Addresses {
		if infra.AddrEqual(a, addr) {
			return
		}
	}

	wrapped := wrapAddr(addr)
	c.Addresses = append(c.Addresses, wrapped)
	c.queue = c.queue.Push(wrapped)
}

func (c *Contact) Remove(addr infra.Addr) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i, a := range c.Addresses {
		if infra.AddrEqual(a, addr) {
			c.Addresses = append(c.Addresses[:i], c.Addresses[i+1:]...)
			return
		}
	}
}
