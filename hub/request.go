package hub

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"io"
)

// Request is a request handler sent to the port handler
type Request struct {
	caller     *id.Identity
	response   chan bool
	connection chan Conn
	query      string
}

func NewRequest(caller *id.Identity, query string) *Request {
	return &Request{
		caller:     caller,
		query:      query,
		response:   make(chan bool, 1),
		connection: make(chan Conn, 1),
	}
}

// Caller returns the auth.Identity of the requesting party
func (req *Request) Caller() *id.Identity {
	return req.caller
}

// Query returns the name of the port that the request was sent to
func (req *Request) Query() string {
	return req.query
}

// Reject rejects the request
func (req *Request) Reject() {
	defer close(req.response)
	req.response <- false
}

// Accept accepts the request
func (req *Request) Accept() io.ReadWriteCloser {
	defer close(req.response)
	req.response <- true
	return <-req.connection
}
