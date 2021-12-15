package contacts

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/logfmt"
	"sync"
	"time"
)

const defaultAddressValidity = time.Hour * 24 * 30

type Contact struct {
	identity  id.Identity
	alias     string
	mu        sync.Mutex
	Addresses []Addr
}

func NewContact(identity id.Identity) *Contact {
	return &Contact{
		identity:  identity,
		Addresses: make([]Addr, 0),
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

func (c *Contact) Add(addr infra.Addr) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, a := range c.Addresses {
		if infra.AddrEqual(a, addr) {
			return
		}
	}
	c.Addresses = append(c.Addresses, Addr{Addr: addr, ExpiresAt: time.Now().Add(defaultAddressValidity)})
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
