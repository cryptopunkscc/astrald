package tracker

import (
	"database/sql"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
	"time"
)

// Add adds an endpoint to the identity. If the endpoint already exists, its expiry time will be replaced.
func (tracker *CoreTracker) Add(identity id.Identity, e net.Endpoint, expiresAt time.Time) error {
	idHex := identity.PublicKeyHex()

	// repack the endpoint to validate it and get a concrete type if possible
	e, err := tracker.unpacker.Unpack(e.Network(), e.Pack())
	if err != nil {
		return err
	}

	return tracker.db.TxDo(func(tx *sql.Tx) (err error) {
		_, err = tx.Exec(
			queryDeleteAddr,
			idHex,
			e.Network(),
			string(e.Pack()),
		)
		if err != nil {
			return
		}

		_, err = tx.Exec(
			queryInsert,
			idHex,
			e.Network(),
			string(e.Pack()),
			expiresAt,
		)
		if err != nil {
			return
		}
		tracker.events.Emit(EventNewAddr{
			NodeID: identity,
			Addr: &Addr{
				Endpoint:  e,
				ExpiresAt: expiresAt,
			}})

		_, err = tx.Exec(queryPurge)

		return
	})
}
