package contacts

import "github.com/cryptopunkscc/astrald/node/db"

const dbTableName = "contacts"
const dbTableFields = "nodeID string, alias string"

const queryInsertNodeID = "INSERT INTO contacts (nodeID) VALUES ($1)"
const queryDeleteByNodeID = "DELETE FROM contacts WHERE nodeID = $1"
const queryCountContacts = "SELECT count() FROM contacts"
const querySelectAllContacts = "SELECT nodeID, alias FROM contacts"
const querySelectAliasByID = "SELECT alias FROM contacts WHERE nodeID = $1"
const querySelectIDByAlias = "SELECT nodeID FROM contacts WHERE alias = $1"
const queryUpdateContact = "UPDATE contacts SET alias = $1 WHERE nodeID = $2"

func InitDatabase(db *db.Database) error {
	return db.CreateTable(dbTableName, dbTableFields)
}
