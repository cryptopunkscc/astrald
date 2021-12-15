package contacts

import (
	"bytes"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/enc"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/logfmt"
	"github.com/cryptopunkscc/astrald/nodeinfo"
	"io"
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

func (c Contact) Pack() []byte {
	buf := &bytes.Buffer{}
	_ = writeInfo(buf, &c)
	return buf.Bytes()
}

func Unpack(data []byte) (*Contact, error) {
	return readInfo(bytes.NewReader(data))
}

func writeInfo(w io.Writer, c *Contact) error {
	err := enc.WriteIdentity(w, c.Identity())
	if err != nil {
		return err
	}

	err = enc.WriteL8String(w, c.Alias())
	if err != nil {
		return err
	}

	addrs := c.Addresses[:]
	if len(addrs) > 255 {
		addrs = addrs[:255]
	}

	err = enc.Write(w, uint8(len(addrs)))
	if err != nil {
		return err
	}

	for _, addr := range addrs {
		if err := nodeinfo.WriteAddr(w, addr); err != nil {
			return nil
		}
	}

	return nil
}

func readInfo(r io.Reader) (*Contact, error) {
	identity, err := enc.ReadIdentity(r)
	if err != nil {
		return nil, err
	}

	alias, err := enc.ReadL8String(r)
	if err != nil {
		return nil, err
	}

	count, err := enc.ReadUint8(r)
	if err != nil {
		return nil, err
	}

	addrs := make([]Addr, 0, count)
	for i := 0; i < int(count); i++ {
		addr, err := nodeinfo.ReadAddr(r)
		if err != nil {
			return nil, err
		}
		addrs = append(addrs, Addr{Addr: addr, ExpiresAt: time.Now().Add(defaultAddressValidity)})
	}

	return &Contact{
		identity:  identity,
		alias:     alias,
		Addresses: addrs,
	}, nil
}
