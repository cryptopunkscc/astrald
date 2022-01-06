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

func (c *Contact) FollowAddr(ctx context.Context, onlyNew bool) <-chan *Addr {
	var ch chan *Addr

	if onlyNew {
		ch = make(chan *Addr)
	} else {
		ch = make(chan *Addr, len(c.Addresses))
		for _, a := range c.Addresses {
			ch <- a
		}
	}

	go func() {
		defer close(ch)
		for i := range c.queue.Follow(ctx) {
			ch <- i.(*Addr)
		}
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
