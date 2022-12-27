package contacts

import (
	"github.com/cryptopunkscc/astrald/infra"
	"time"
)

func (c *Contact) Add(addr infra.Addr, expiresAt time.Time) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.add(addr, expiresAt)
}

func (c *Contact) add(addr infra.Addr, expiresAt time.Time) error {
	for _, a := range c.addresses {
		if infra.AddrEqual(a, addr) {
			if a.ExpiresAt.Before(expiresAt) {
				a.ExpiresAt = expiresAt
			}
			return nil
		}
	}

	a := &Addr{Addr: addr, ExpiresAt: expiresAt}
	if expiresAt.IsZero() {
		a.ExpiresAt = time.Now().Add(defaultAddressValidity)
	}
	c.addresses = append(c.addresses, a)
	c.queue = c.queue.Push(a)

	return nil
}
