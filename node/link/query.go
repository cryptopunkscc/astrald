package link

import "github.com/cryptopunkscc/astrald/astral/link"

type Query struct {
	link *Link
	*link.Query
}

func (query *Query) Accept() (*Conn, error) {
	conn, err := query.Query.Accept()

	if conn != nil {
		go func() {
			query.link.Activity.Add(1)
			defer query.link.Activity.Done()
			<-conn.Wait()
		}()
	}

	return wrapConn(conn), err
}

func (query *Query) Link() *Link {
	return query.link
}
