package link

import (
	"github.com/cryptopunkscc/astrald/cslq"
	"io"
	"sync/atomic"
	"time"
)

const DefaultQueryTimeout = 30 * time.Second

type Query struct {
	query     string
	localPort int
	reader    io.Reader
	writer    io.WriteCloser
	link      *Link
	done      atomic.Bool
}

func (query *Query) Query() string {
	return query.query
}

func (query *Query) Accept() (*Conn, error) {
	if query.done.CompareAndSwap(false, true) {
		cslq.Encode(query.writer, "s", query.localPort)

		conn := &Conn{
			localPort: query.localPort,
			query:     query.query,
			reader:    query.reader,
			writer:    query.writer,
			link:      query.link,
			done:      make(chan struct{}),
		}

		if port, ok := conn.reader.(*PortReader); ok {
			port.SetErrorHandler(func(err error) {
				if err == ErrBufferOverflow {
					conn.closeWithError(err)
				}
			})
		}

		query.link.add(conn)
		query.link.Events().Emit(EventConnEstablished{Conn: conn})

		return conn, nil
	}
	return nil, ErrQueryFinished
}

func (query *Query) Reject() error {
	if query.done.CompareAndSwap(false, true) {
		query.writer.Close()
		query.link.mux.Unbind(query.localPort)
		return nil
	}

	return ErrQueryFinished
}

func (query *Query) Link() *Link {
	return query.link
}
