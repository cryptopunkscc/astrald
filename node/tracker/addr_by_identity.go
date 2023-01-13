package tracker

import (
	"database/sql"
	"github.com/cryptopunkscc/astrald/auth/id"
)

// AddrByIdentity returns all addresses for the provided identity.
func (tracker *Tracker) AddrByIdentity(identity id.Identity) ([]Addr, error) {
	dbRows, err := tracker.dbAddrByIdentity(identity)
	if err != nil {
		return nil, err
	}

	res := make([]Addr, 0, len(dbRows))
	for _, dbRow := range dbRows {
		addr, err := tracker.unpacker.Unpack(dbRow.Network, []byte(dbRow.Address))
		if err != nil {
			return nil, err
		}
		res = append(res, Addr{
			Addr:      addr,
			ExpiresAt: dbRow.ExpiresAt,
		})
	}

	return res, nil
}

func (tracker *Tracker) dbAddrByIdentity(identity id.Identity) ([]dbAddr, error) {
	var dbRows = make([]dbAddr, 0)
	var idHex = identity.PublicKeyHex()

	err := tracker.db.TxDo(func(tx *sql.Tx) error {
		results, err := tx.Query(queryUnexpiredAddrsByIdentity, idHex)
		if err != nil {
			return err
		}

		var row dbAddr
		for results.Next() {
			if err := results.Scan(&row.NodeID, &row.Network, &row.Address, &row.ExpiresAt); err != nil {
				return err
			}
			dbRows = append(dbRows, row)
		}

		return nil
	})

	return dbRows, err
}
