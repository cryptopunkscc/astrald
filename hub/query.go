package hub

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/link"
)

// Query is a request handler sent to the port handler
type Query struct {
	query      string
	link       *link.Link
	response   chan bool
	connection chan *Conn
}

func NewQuery(query string, link *link.Link) *Query {
	return &Query{
		query:      query,
		link:       link,
		response:   make(chan bool, 1),
		connection: make(chan *Conn, 1),
	}
}

// Link returns the link from which the query is coming or nil in case of local queries
func (query *Query) Link() *link.Link {
	return query.link
}

// RemoteIdentity returns the remote identity of the caller
func (query *Query) RemoteIdentity() id.Identity {
	if query.link != nil {
		return query.link.RemoteIdentity()
	}

	return id.Identity{}
}

// IsLocal returns true if query is local
func (query *Query) IsLocal() bool {
	return query.link == nil
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
func (query *Query) Accept() *Conn {
	defer close(query.response)
	query.response <- true
	return <-query.connection
}
