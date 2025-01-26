package apphost

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"net"
)

type PendingQuery struct {
	conn net.Conn
	info *apphost.QueryInfo
}

func (q *PendingQuery) Query() string {
	return string(q.info.Query)
}

func (q *PendingQuery) Caller() *astral.Identity {
	return q.info.Caller
}

func (q *PendingQuery) Reject() (err error) {
	_, err = astral.Uint8(1).WriteTo(q.conn)
	q.conn.Close()

	return
}

func (q *PendingQuery) Accept() (conn *Conn, err error) {
	_, err = astral.Uint8(0).WriteTo(q.conn)
	if err != nil {
		return
	}

	conn = &Conn{
		Conn:     q.conn,
		remoteID: q.info.Caller,
		query:    string(q.info.Query),
	}

	return conn, nil
}

func (q *PendingQuery) Close() error {
	return q.conn.Close()
}
