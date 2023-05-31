package services

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/link"
	"sync/atomic"
)

const SourceNetwork = "network"
const SourceLocal = "local"

// Query is a request handler sent to the port handler
type Query struct {
	query    string
	source   string
	link     *link.Link
	response chan bool
	handled  atomic.Bool
	conn     *Conn
	err      error
	remoteID id.Identity
}

func newQuery(query string, link *link.Link, remoteID id.Identity, conn *Conn) *Query {
	return &Query{
		query:    query,
		link:     link,
		remoteID: remoteID,
		conn:     conn,
		response: make(chan bool, 1),
	}
}

// RemoteIdentity returns the remote identity of the caller
func (query *Query) RemoteIdentity() id.Identity {
	return query.remoteID
}

func (query *Query) Source() string {
	if query.link != nil {
		return SourceNetwork
	}
	return SourceLocal
}

// Query returns query string
func (query *Query) Query() string {
	return query.query
}

// Reject rejects the query
func (query *Query) Reject() error {
	if query.handled.CompareAndSwap(false, true) {
		query.response <- false
		query.err = ErrQueryHandled
		return nil
	}
	return query.err
}

// Accept accepts the query
func (query *Query) Accept() (*Conn, error) {
	if query.handled.CompareAndSwap(false, true) {
		query.response <- true
		query.err = ErrQueryHandled
		return query.conn, nil
	}
	return nil, query.err
}

func (query *Query) cancel(e error) error {
	if query.handled.CompareAndSwap(false, true) {
		query.response <- false
		query.err = e
		return nil
	}
	return query.err
}
