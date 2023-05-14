package astral

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mod/apphost/proto"
	"net"
)

type QueryData struct {
	conn     net.Conn
	query    string
	remoteID id.Identity
}

func (q *QueryData) Query() string {
	return q.query
}

func (q *QueryData) RemoteIdentity() id.Identity {
	return q.remoteID
}

func (q *QueryData) Reject() error {
	defer q.conn.Close()
	return cslq.Encode(q.conn, "c", 1)
}

func (q *QueryData) Accept() (*Conn, error) {
	conn := &Conn{
		Conn:     q.conn,
		remoteID: q.remoteID,
		query:    q.query,
	}

	if err := cslq.Encode(q.conn, "c", proto.Success); err != nil {
		return nil, err
	}

	return conn, nil
}
