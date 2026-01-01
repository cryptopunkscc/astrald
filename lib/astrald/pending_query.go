package astrald

import (
	"net"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	libapphost "github.com/cryptopunkscc/astrald/lib/apphost"
	"github.com/cryptopunkscc/astrald/mod/apphost"
)

type PendingQuery struct {
	conn  net.Conn
	query *astral.Query
}

// Accept accepts the query and returns a new connection
func (q *PendingQuery) Accept() (conn *libapphost.Conn) {
	// send ack
	_ = channel.NewBinaryWriter(q.conn).Write(&astral.Ack{})

	conn = libapphost.NewConn(
		q.conn,
		q.query.Caller,
		q.query.Target,
		q.query.Query,
		q.query.Nonce,
	)

	return conn
}

// AcceptChannel accepts the query and returns a new channel
func (q *PendingQuery) AcceptChannel() *channel.Channel {
	return channel.New(q.Accept())
}

// Reject rejects the query with the default error code
func (q *PendingQuery) Reject() (err error) {
	return q.RejectWithCode(astral.CodeRejected)
}

// RejectWithCode rejects the query with the given code
func (q *PendingQuery) RejectWithCode(code int) (err error) {
	defer q.conn.Close()
	return channel.NewBinaryWriter(q.conn).Write(&apphost.QueryRejectedMsg{Code: astral.Uint8(code)})
}

// Skip responds with a "route not found" error
func (q *PendingQuery) Skip() error {
	defer q.conn.Close()
	return channel.NewBinaryWriter(q.conn).Write(&apphost.ErrorMsg{Code: apphost.ErrCodeRouteNotFound})
}

// Close closes the connection without responding
func (q *PendingQuery) Close() error {
	return q.conn.Close()
}

func (q *PendingQuery) Nonce() astral.Nonce { return q.query.Nonce }

func (q *PendingQuery) Caller() *astral.Identity {
	return q.query.Caller
}

func (q *PendingQuery) Target() *astral.Identity {
	return q.query.Target
}

func (q *PendingQuery) Query() string {
	return q.query.Query
}
