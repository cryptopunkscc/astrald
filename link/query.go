package link

import (
	"encoding/binary"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mux"
)

type Query struct {
	query        string
	inputStream  *mux.InputStream
	outputStream *mux.OutputStream
	link         *Link
}

// Accept the query
func (query *Query) Accept() (*Conn, error) {
	err := binary.Write(query.outputStream, binary.BigEndian, uint16(query.inputStream.ID()))
	if err != nil {
		return nil, err
	}

	conn := NewConn(query.inputStream, query.outputStream, false, query.query)
	conn.Attach(query.link)

	return conn, nil
}

// Reject the query
func (query Query) Reject() error {
	query.inputStream.Close()

	return query.outputStream.Close()
}

// Caller returns the identity of the caller
func (query Query) Caller() id.Identity {
	return query.link.RemoteIdentity()
}

// Query returns the query string
func (query Query) String() string {
	return query.query
}
