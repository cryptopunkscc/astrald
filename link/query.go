package link

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mux"
)

type Query struct {
	query string
	in    *mux.InputStream
	out   *mux.OutputStream
	link  *Link
}

// Accept the query
func (query *Query) Accept() (*Conn, error) {
	err := cslq.Encode(query.out, "s", query.in.ID())
	if err != nil {
		return nil, err
	}

	conn := newConn(query.in, query.out, false, query.query)
	conn.Attach(query.link)

	return conn, nil
}

// Reject the query
func (query Query) Reject() error {
	query.out.Close()
	
	return query.in.Discard()
}

// Caller returns the identity of the caller
func (query Query) Caller() id.Identity {
	return query.link.RemoteIdentity()
}

// Query returns the query string
func (query Query) String() string {
	return query.query
}
