package tracker

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	_log "github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/db"
)

const tableName = "addrs"
const tableFields = "nodeID string, network string, address string, expiresAt time"

const queryPurge = "DELETE FROM addrs WHERE expiresAt <= now()"
const queryDeleteAddr = "DELETE FROM addrs WHERE nodeID = $1 AND network = $2 AND address = $3"
const queryDeleteByNodeID = "DELETE FROM addrs WHERE nodeID = $1"
const queryInsert = "INSERT INTO addrs VALUES ($1, $2, $3, $4)"
const queryUnexpiredAddrsByIdentity = "SELECT * FROM addrs WHERE nodeID = $1 AND expiresAt > now()"
const queryUniqueIDs = "SELECT DISTINCT nodeID FROM addrs"

var log = _log.Tag("tracker")

type AddrUnpacker interface {
	Unpack(string, []byte) (net.Endpoint, error)
}

type EventNewAddr struct {
	NodeID id.Identity
	*Addr
}

func InitDatabase(db *db.Database) error {
	return db.CreateTable(tableName, tableFields)
}
