package tracker

import (
	"database/sql"
	"github.com/cryptopunkscc/astrald/auth/id"
)

// Identities returns a list of all tracked identities
func (tracker *CoreTracker) Identities() ([]id.Identity, error) {
	ids := make([]id.Identity, 0)

	err := tracker.db.TxDo(func(tx *sql.Tx) error {
		rows, err := tx.Query(queryUniqueIDs)
		if err != nil {
			return err
		}

		for rows.Next() {
			var hex string
			if err := rows.Scan(&hex); err != nil {
				return err
			}

			nodeID, err := id.ParsePublicKeyHex(hex)
			if err != nil {
				continue
			}

			ids = append(ids, nodeID)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return ids, nil
}
