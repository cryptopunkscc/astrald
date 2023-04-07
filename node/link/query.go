package link

import (
	"errors"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mux"
	"sync/atomic"
	"time"
)

var ErrQueryFinished = errors.New("query finished")

const DefaultQueryTimeout = 30 * time.Second

type Query struct {
	query     string
	localPort int
	input     *mux.FrameReader
	output    *mux.FrameWriter
	link      *Link
	done      atomic.Bool
}

func (query *Query) Query() string {
	return query.query
}

func (query *Query) Accept() (*Conn, error) {
	if query.done.CompareAndSwap(false, true) {
		cslq.Encode(query.output, "s", query.localPort)

		conn := &Conn{
			localPort: query.localPort,
			query:     query.query,
			reader:    query.input,
			writer:    query.output,
			link:      query.link,
			done:      make(chan struct{}),
		}

		query.link.add(conn)
		query.link.Events().Emit(EventConnEstablished{Conn: conn})

		return conn, nil
	}
	return nil, ErrQueryFinished
}

func (query *Query) Reject() error {
	if query.done.CompareAndSwap(false, true) {
		query.output.Close()
		query.link.mux.Unbind(query.localPort)
		return nil
	}

	return ErrQueryFinished
}

func (query *Query) Link() *Link {
	return query.link
}
