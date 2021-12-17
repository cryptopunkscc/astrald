package contacts

import (
	"encoding/json"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/logfmt"
	"github.com/cryptopunkscc/astrald/nodeinfo"
	"github.com/cryptopunkscc/astrald/storage"
	"log"
	"sync"
)

const contactsKey = "contacts.json"

var _ Resolver = &Manager{}

type Manager struct {
	contacts map[string]*Contact
	store    storage.Store
	mu       sync.Mutex
}

func New(store storage.Store) *Manager {
	c := &Manager{
		store:    store,
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

func (m *Manager) Find(identity id.Identity, create bool) *Contact {
	m.mu.Lock()
	defer m.mu.Unlock()

	hex := identity.PublicKeyHex()

	if c, found := m.contacts[hex]; found {
		return c
	}

	if create {
		m.contacts[hex] = NewContact(identity)
		return m.contacts[hex]
	}

	return nil
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
		c.Add(a)
	}

	m.Save()

	return nil
}

// Lookup returns all known addresses of a node via channel
func (m *Manager) Lookup(nodeID id.Identity) <-chan infra.Addr {
	m.mu.Lock()
	defer m.mu.Unlock()

	c := m.contacts[nodeID.PublicKeyHex()]
	if c == nil {
		ch := make(chan infra.Addr)
		close(ch)
		return ch
	}

	// convert to channel
	ch := make(chan infra.Addr, len(c.Addresses))
	for _, addr := range c.Addresses {
		ch <- addr.Addr
	}
	close(ch)

	return ch
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
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}

	return m.store.StoreBytes(contactsKey, data)
}

func (m *Manager) load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	bytes, err := m.store.LoadBytes(contactsKey)
	if err != nil {
		return err
	}

	return json.Unmarshal(bytes, m)
}
