package contacts

import "github.com/cryptopunkscc/astrald/auth/id"

func (m *Manager) Find(identity id.Identity, create bool) *Contact {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.find(identity, create)
}

func (m *Manager) find(identity id.Identity, create bool) *Contact {
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
