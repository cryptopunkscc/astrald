package contacts

import (
	"database/sql"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/logfmt"
	"github.com/cryptopunkscc/astrald/node/db"
)

type Contact struct {
	db       *db.Database
	identity id.Identity
	alias    string
}

type dbContact struct {
	NodeID string
	Alias  string
}

func (c *Contact) Identity() id.Identity {
	return c.identity
}

func (c *Contact) Alias() string {
	return c.alias
}

func (c *Contact) SetAlias(alias string) (err error) {
	prev := c.alias
	c.alias = alias
	if err = c.save(); err != nil {
		c.alias = prev
	}
	return
}

func (c *Contact) DisplayName() string {
	if c.alias != "" {
		return c.alias
	}

	return logfmt.ID(c.identity)
}

func (c *Contact) save() error {
	return c.db.TxDo(func(tx *sql.Tx) error {
		_, err := tx.Exec(queryUpdateContact, c.alias, c.identity.PublicKeyHex())

		return err
	})
}

func (dbc *dbContact) toContact(db *db.Database) (*Contact, error) {
	var err error
	var c = &Contact{
		db: db,
	}

	c.alias = dbc.Alias
	c.identity, err = id.ParsePublicKeyHex(dbc.NodeID)
	if err != nil {
		return nil, err
	}
	return c, nil
}
