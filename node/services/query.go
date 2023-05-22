package services

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/link"
	"sync"
)

// Query is a request handler sent to the port handler
type Query struct {
	query      string
	link       *link.Link
	response   chan bool
	connection chan *Conn
	err        error
	mu         sync.Mutex
	remoteID   id.Identity
}

func NewQuery(query string, link *link.Link, remoteID id.Identity) *Query {
	return &Query{
		query:      query,
		link:       link,
		remoteID:   remoteID,
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
	if !query.remoteID.IsZero() {
		return query.remoteID
	}

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
func (query *Query) Reject() error {
	query.mu.Lock()
	defer query.mu.Unlock()

	query.response <- false
	close(query.response)

	return query.err
}

// Accept accepts the query
func (query *Query) Accept() (*Conn, error) {
	query.mu.Lock()
	defer query.mu.Unlock()

	if query.err != nil {
		return nil, query.err
	}

	query.response <- true
	close(query.response)

	return <-query.connection, nil
}

func (query *Query) setError(e error) {
	query.mu.Lock()
	defer query.mu.Unlock()

	query.err = e
}
