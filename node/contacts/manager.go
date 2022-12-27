package contacts

import (
	"bytes"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/logfmt"
	"github.com/cryptopunkscc/astrald/nodeinfo"
	"github.com/cryptopunkscc/astrald/storage"
	"log"
	"sync"
	"time"
)

const contactsKey = "contacts.dat"

type AddrUnpacker interface {
	Unpack(string, []byte) (infra.Addr, error)
}

type Manager struct {
	unpacker AddrUnpacker
	contacts map[string]*Contact
	store    storage.Store
	mu       sync.Mutex
}

func New(store storage.Store, unpacker AddrUnpacker) *Manager {
	c := &Manager{
		store:    store,
		unpacker: unpacker,
		contacts: make(map[string]*Contact),
	}

	if err := c.load(); err != nil {
		log.Println("load contacts error:", err)
	}

	return c
}

func (m *Manager) DisplayName(identity id.Identity) string {
	if identity.IsZero() {
		return "unknown"
	}

	if c := m.Find(identity, false); c != nil {
		return c.DisplayName()
	}

	return logfmt.ID(identity)
}

func (m *Manager) FindByAlias(alias string) (*Contact, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, c := range m.contacts {
		if c.Alias() == alias {
			return c, true
		}
	}

	return nil, false
}

func (m *Manager) Forget(identity id.Identity) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	hex := identity.PublicKeyHex()

	if _, found := m.contacts[hex]; !found {
		return errors.New("not found")
	}

	delete(m.contacts, hex)

	return m.Save()
}

func (m *Manager) ResolveIdentity(str string) (id.Identity, error) {
	if id, err := id.ParsePublicKeyHex(str); err == nil {
		return id, nil
	}

	if c, found := m.FindByAlias(str); found {
		return c.Identity(), nil
	}

	return id.Identity{}, errors.New("unknown identity")
}

func (m *Manager) AddNodeInfo(info *nodeinfo.NodeInfo) error {
	c := m.Find(info.Identity, true)
	if c.Alias() == "" {
		c.SetAlias(info.Alias)
	}

	for _, a := range info.Addresses {
		c.Add(a, time.Time{})
	}

	m.Save()

	return nil
}

// Lookup returns all known addresses of a node via channel
func (m *Manager) Lookup(nodeID id.Identity) (<-chan *Addr, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	c, ok := m.contacts[nodeID.PublicKeyHex()]
	if !ok {
		return nil, errors.New("unknown identity")
	}

	// convert to channel
	ch := make(chan *Addr, len(c.addresses))
	for _, addr := range c.addresses {
		ch <- addr
	}
	close(ch)

	return ch, nil
}

// All returns a closed channel populated with all contacts
func (m *Manager) All() <-chan *Contact {
	m.mu.Lock()
	defer m.mu.Unlock()

	ch := make(chan *Contact, len(m.contacts))
	for _, c := range m.contacts {
		ch <- c
	}
	close(ch)

	return ch
}

func (m *Manager) Save() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var data = &bytes.Buffer{}
	var enc = cslq.NewEncoder(data)

	enc.Encode("l", len(m.contacts))

	for _, contact := range m.contacts {
		enc.Encode("v [c]c s", contact.identity, contact.alias, len(contact.addresses))

		for _, addr := range contact.addresses {
			enc.Encode("[c]c [c]c q", addr.Network(), addr.Pack(), addr.ExpiresAt.Unix())
		}
	}

	return m.store.StoreBytes(contactsKey, data.Bytes())
}

func (m *Manager) load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := m.store.LoadBytes(contactsKey)
	if err != nil {
		return err
	}

	var dec = cslq.NewDecoder(bytes.NewReader(data))

	var contactLen int

	if err := dec.Decode("l", &contactLen); err != nil {
		return err
	}
	for i := 0; i < contactLen; i++ {
		var identity id.Identity
		var alias string
		var addrLen int

		if err := dec.Decode("v [c]c s", &identity, &alias, &addrLen); err != nil {
			return err
		}
		c := m.find(identity, true)
		c.alias = alias

		for j := 0; j < addrLen; j++ {
			var network string
			var address []byte
			var expiresAt int64

			if err := dec.Decode("[c]c [c]c q", &network, &address, &expiresAt); err != nil {
				return err
			}

			addr, err := m.unpacker.Unpack(network, address)
			if err != nil {
				return err
			}
			c.Add(addr, time.Unix(expiresAt, 0))
		}
	}

	return nil
}
