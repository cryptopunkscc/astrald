package contacts

import (
	"encoding/json"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra/bt"
	"github.com/cryptopunkscc/astrald/infra/gw"
	"github.com/cryptopunkscc/astrald/infra/inet"
	"github.com/cryptopunkscc/astrald/infra/tor"
	"github.com/cryptopunkscc/astrald/sig"
	"time"
)

type jsonContact struct {
	Identity  string
	Alias     string
	Addresses []*Addr
}

type jsonAddr struct {
	Network   string
	Address   string
	ExpiresAt time.Time
}

func (c *Contact) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonContact{
		Identity:  c.identity.String(),
		Alias:     c.alias,
		Addresses: c.Addresses,
	})
}

func (c *Contact) UnmarshalJSON(data []byte) error {
	var v jsonContact
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	if i, err := id.ParsePublicKeyHex(v.Identity); err != nil {
		return err
	} else {
		c.identity = i
	}

	c.queue = &sig.Queue{}
	c.alias = v.Alias
	c.Addresses = v.Addresses

	return nil
}

func (m *Manager) MarshalJSON() ([]byte, error) {
	list := make([]*Contact, 0, len(m.contacts))
	for _, i := range m.contacts {
		list = append(list, i)
	}
	return json.Marshal(list)
}

func (m *Manager) UnmarshalJSON(data []byte) error {
	var list []*Contact

	if err := json.Unmarshal(data, &list); err != nil {
		return err
	}

	m.contacts = make(map[string]*Contact, len(list))
	for _, i := range list {
		m.contacts[i.identity.PublicKeyHex()] = i
	}

	return nil
}

func (a *Addr) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonAddr{
		Network:   a.Network(),
		Address:   a.String(),
		ExpiresAt: a.ExpiresAt,
	})
}

func (a *Addr) UnmarshalJSON(data []byte) error {
	var ja jsonAddr
	var err error

	if err = json.Unmarshal(data, &ja); err != nil {
		return err
	}

	a.ExpiresAt = ja.ExpiresAt

	switch ja.Network {
	case tor.NetworkName:
		if a.Addr, err = tor.Parse(ja.Address); err != nil {
			return err
		}
	case inet.NetworkName:
		if a.Addr, err = inet.Parse(ja.Address); err != nil {
			return err
		}
	case gw.NetworkName:
		if a.Addr, err = gw.Parse(ja.Address); err != nil {
			return err
		}
	case bt.NetworkName:
		if a.Addr, err = bt.Parse(ja.Address); err != nil {
			return err
		}
	default:
		return errors.New("unsupported network")
	}

	return nil
}
