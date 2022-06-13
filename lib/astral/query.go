package astral

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"net"
)

type Query struct {
	conn     net.Conn
	query    string
	remoteID id.Identity
}

func (q *Query) Query() string {
	return q.query
}

func (q *Query) RemoteIdentity() id.Identity {
	return q.remoteID
}

func (q *Query) Reject() error {
	defer q.conn.Close()
	return cslq.Encode(q.conn, "c", 1)
}

func (q *Query) Accept() (*Conn, error) {
	conn := &Conn{
		Conn:     q.conn,
		remoteID: q.remoteID,
		query:    q.query,
	}

	if err := cslq.Encode(q.conn, "c", 0); err != nil {
		return nil, err
	}

	return conn, nil
}
