package contacts

import (
	"database/sql"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/db"
	"github.com/cryptopunkscc/astrald/node/event"
	"sync"
)

type Manager struct {
	db     *db.Database
	mu     sync.Mutex
	events event.Queue
}

func New(db *db.Database, eventParent *event.Queue) (*Manager, error) {
	var c = &Manager{
		db: db,
	}

	c.events.SetParent(eventParent)

	return c, nil
}

func (m *Manager) DisplayName(nodeID id.Identity) string {
	if nodeID.IsZero() {
		return "unknown"
	}

	if contact, err := m.Find(nodeID); err == nil {
		return contact.DisplayName()
	}

	return nodeID.Fingerprint()
}

func (m *Manager) Find(nodeID id.Identity) (c *Contact, err error) {
	err = m.db.TxDo(func(tx *sql.Tx) error {
		row := tx.QueryRow(querySelectAliasByID, nodeID.PublicKeyHex())
		res := &Contact{
			db:       m.db,
			identity: nodeID,
		}

		if err = row.Scan(&res.alias); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return errors.New("not found")
			}
			return err
		}

		c = res
		return nil
	})
	return
}

func (m *Manager) FindOrCreate(nodeID id.Identity) (c *Contact, err error) {
	if c, err = m.Find(nodeID); err == nil {
		return
	}

	err = m.db.TxDo(func(tx *sql.Tx) error {
		_, err := tx.Exec(queryInsertNodeID, nodeID.PublicKeyHex())
		if err != nil {
			return err
		}

		log.Info("%s%s%s added to contacts", log.Cyan(), nodeID.Fingerprint(), log.Reset())

		m.events.Emit(EventContactAdded{nodeID})

		c = &Contact{identity: nodeID, db: m.db}
		return nil
	})
	return
}

func (m *Manager) FindByAlias(alias string) (c *Contact, found bool) {
	m.db.TxDo(func(tx *sql.Tx) error {
		res, err := tx.Query(querySelectIDByAlias, alias)
		if err != nil {
			return err
		}
		if !res.Next() {
			return errors.New("not found")
		}

		var nodeIDHex string
		err = res.Scan(&nodeIDHex)
		if err != nil {
			return err
		}

		nodeID, err := id.ParsePublicKeyHex(nodeIDHex)
		if err != nil {
			return err
		}

		c = &Contact{
			db:       m.db,
			identity: nodeID,
			alias:    alias,
		}
		found = true

		return nil
	})
	return
}

func (m *Manager) Delete(identity id.Identity) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.db.TxDo(func(tx *sql.Tx) error {
		res, err := tx.Exec(queryDeleteByNodeID, identity.PublicKeyHex())
		if err != nil {
			return err
		}
		if i, err := res.RowsAffected(); i == 0 {
			if err != nil {
				return err
			}
			return errors.New("not found")
		}
		return nil
	})
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

// All returns a closed channel populated with all contacts
func (m *Manager) All() <-chan *Contact {
	var ch chan *Contact
	m.db.TxDo(func(tx *sql.Tx) error {
		rows, err := tx.Query(queryCountContacts)
		if err != nil {
			return err
		}
		if !rows.Next() {
			return errors.New("count returned no rows")
		}
		var count int
		if err := rows.Scan(&count); err != nil {
			return err
		}

		ch = make(chan *Contact, count)
		defer close(ch)

		rows, err = tx.Query(querySelectAllContacts)
		if err != nil {
			return err
		}

		for rows.Next() {
			dbc := &dbContact{}
			if err := rows.Scan(&dbc.NodeID, &dbc.Alias); err != nil {
				return err
			}

			c, err := dbc.toContact(m.db)
			if err != nil {
				return err
			}
			ch <- c
		}

		return nil
	})
	return ch
}
