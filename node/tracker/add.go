package tracker

import (
	"database/sql"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
	"time"
)

// Add adds an address to the identity. If the address already exists, its expiry time will be replaced.
func (tracker *Tracker) Add(identity id.Identity, addr infra.Addr, expiresAt time.Time) error {
	idHex := identity.PublicKeyHex()

	// repack the address to validate it and get a concrete type if possible
	addr, err := tracker.unpacker.Unpack(addr.Network(), addr.Pack())
	if err != nil {
		return err
	}

	return tracker.db.TxDo(func(tx *sql.Tx) (err error) {
		_, err = tx.Exec(
			queryDeleteAddr,
			idHex,
			addr.Network(),
			string(addr.Pack()),
		)
		if err != nil {
			return
		}

		_, err = tx.Exec(
			queryInsert,
			idHex,
			addr.Network(),
			string(addr.Pack()),
			expiresAt,
		)
		if err != nil {
			return
		}
		tracker.events.Emit(EventNewAddr{
			NodeID: identity,
			Addr: &Addr{
				Addr:      addr,
				ExpiresAt: expiresAt,
			}})

		_, err = tx.Exec(queryPurge)

		return
	})
}
