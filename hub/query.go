package hub

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"io"
)

// Query is a request handler sent to the port handler
type Query struct {
	caller     id.Identity
	response   chan bool
	connection chan Conn
	query      string
}

func NewQuery(caller id.Identity, query string) *Query {
	return &Query{
		caller:     caller,
		query:      query,
		response:   make(chan bool, 1),
		connection: make(chan Conn, 1),
	}
}

// Caller returns the Identity of the calller
func (query *Query) Caller() id.Identity {
	return query.caller
}

// Query returns query string
func (query *Query) Query() string {
	return query.query
}

// Reject rejects the query
func (query *Query) Reject() {
	defer close(query.response)
	query.response <- false
}

// Accept accepts the query
func (query *Query) Accept() io.ReadWriteCloser {
	defer close(query.response)
	query.response <- true
	return <-query.connection
}
