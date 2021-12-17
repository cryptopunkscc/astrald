package link

import (
	"github.com/cryptopunkscc/astrald/link"
)

type Query struct {
	link *Link
	*link.Query
}

func (query *Query) Accept() (*Conn, error) {
	rawConn, err := query.Query.Accept()
	if err != nil {
		return nil, err
	}

	conn := wrapConn(rawConn)
	query.link.add(conn)

	return conn, err
}

func (query *Query) Link() *Link {
	return query.link
}
