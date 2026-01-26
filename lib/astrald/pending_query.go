package astrald

import (
	"net"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	libapphost "github.com/cryptopunkscc/astrald/lib/apphost"
	"github.com/cryptopunkscc/astrald/mod/apphost"
)

// OnQueryAccepted is called when a query is accepted. It can be used with a RouterMonitor
// to keep track of all open connections.
var OnQueryAccepted func(conn astral.Conn, query *astral.Query)

type PendingQuery struct {
	conn  net.Conn
	query *astral.Query
}

// Accept accepts the query and returns a new connection
func (pending *PendingQuery) Accept() (conn *libapphost.Conn) {
	// send ack
	_ = channel.NewBinarySender(pending.conn).Send(&astral.Ack{})

	conn = libapphost.NewConn(pending.conn, pending.query, false)

	if OnQueryAccepted != nil {
		OnQueryAccepted(conn, pending.query)
	}

	return conn
}

// AcceptChannel accepts the query and returns a new channel
func (pending *PendingQuery) AcceptChannel(cfg ...channel.ConfigFunc) *channel.Channel {
	return channel.New(pending.Accept(), cfg...)
}

// Reject rejects the query with the default error code
func (pending *PendingQuery) Reject() (err error) {
	return pending.RejectWithCode(astral.CodeRejected)
}

// RejectWithCode rejects the query with the given code
func (pending *PendingQuery) RejectWithCode(code int) (err error) {
	defer pending.conn.Close()
	return channel.NewBinarySender(pending.conn).Send(&apphost.QueryRejectedMsg{Code: astral.Uint8(code)})
}

// Skip responds with a "route not found" error
func (pending *PendingQuery) Skip() error {
	defer pending.conn.Close()
	return channel.NewBinarySender(pending.conn).Send(&apphost.ErrorMsg{Code: apphost.ErrCodeRouteNotFound})
}

// Close closes the connection without responding
func (pending *PendingQuery) Close() error {
	return pending.conn.Close()
}

func (pending *PendingQuery) Nonce() astral.Nonce { return pending.query.Nonce }

func (pending *PendingQuery) Caller() *astral.Identity {
	return pending.query.Caller
}

func (pending *PendingQuery) Target() *astral.Identity {
	return pending.query.Target
}

func (pending *PendingQuery) Query() string {
	return pending.query.Query
}
