package tracker

import (
	"database/sql"
	"github.com/cryptopunkscc/astrald/auth/id"
)

// purge cleans the database of expired addresses
func (tracker *CoreTracker) purge() error {
	return tracker.db.TxDo(func(tx *sql.Tx) error {
		_, err := tx.Exec(queryPurge)
		return err
	})
}

func (tracker *CoreTracker) ForgetIdentity(identity id.Identity) error {
	return tracker.db.TxDo(func(tx *sql.Tx) error {
		_, err := tx.Exec(queryDeleteByNodeID, identity.PublicKeyHex())
		return err
	})
}
