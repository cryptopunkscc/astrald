package link

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth"
	"github.com/cryptopunkscc/astrald/link"
	"github.com/cryptopunkscc/astrald/node/event"
	"github.com/cryptopunkscc/astrald/sig"
	"io"
	"strings"
	"sync"
	"time"
)

const queryChanLen = 16
const defaultIdleTimeout = 60 * time.Minute
const pingInterval = 30 * time.Minute
const pingTimeout = 15 * time.Second

// Link wraps an astral.Link and adds activity tracking
type Link struct {
	sig.Activity
	*link.Link
	queries          chan *Query
	events           event.Queue
	conns            map[*Conn]struct{}
	mu               sync.Mutex
	establishedAt    time.Time
	roundtrip        time.Duration
	idleTimeout      time.Duration
	setIdleTimeoutCh chan time.Duration
	err              error
}

func New(link *link.Link) *Link {
	l := &Link{
		Link:             link,
		queries:          make(chan *Query, queryChanLen),
		conns:            make(map[*Conn]struct{}, 0),
		establishedAt:    time.Now(),
		roundtrip:        999 * time.Second, // assume super slow before first ping
		setIdleTimeoutCh: make(chan time.Duration, 1),
		idleTimeout:      defaultIdleTimeout,
	}

	return l
}

func NewFromConn(conn auth.Conn) *Link {
	return New(link.New(conn))
}

func (l *Link) Query(ctx context.Context, query string) (*Conn, error) {
	if len(query) == 0 {
		return nil, errors.New("empty query")
	}

	// silent queries do not affect activity
	if !(query[0] == '.') {
		l.Activity.Add(1)
		defer l.Activity.Done()
	}

	linkConn, err := l.Link.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	conn := wrapConn(linkConn)
	conn.Attach(l)

	l.events.Emit(EventConnEstablished{conn})

	return conn, nil
}

func (l *Link) Queries() <-chan *Query {
	return l.queries
}

func (l *Link) Err() error {
	return l.err
}

func (l *Link) EstablishedAt() time.Time {
	return l.establishedAt
}

func (l *Link) Subscribe(ctx context.Context) <-chan event.Event {
	return l.events.Subscribe(ctx)
}

func (l *Link) Conns() <-chan *Conn {
	l.mu.Lock()
	defer l.mu.Unlock()

	ch := make(chan *Conn, len(l.conns))
	for l := range l.conns {
		ch <- l
	}
	close(ch)
	return ch
}

func (l *Link) ConnCount() int {
	l.mu.Lock()
	defer l.mu.Unlock()

	return len(l.conns)
}

func (l *Link) SetEventParent(parent *event.Queue) {
	l.events.SetParent(parent)
}

func (l *Link) add(conn *Conn) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// skip duplicates
	if _, found := l.conns[conn]; found {
		return
	}

	l.conns[conn] = struct{}{}
	l.Activity.Add(1)
}

func (l *Link) remove(conn *Conn) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if _, found := l.conns[conn]; !found {
		return errors.New("not found")
	}
	delete(l.conns, conn)
	l.Activity.Done()

	return nil
}

func (l *Link) processQueries(ctx context.Context) error {
	defer close(l.queries)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case query, ok := <-l.Link.Queries():
			if !ok {
				l.err = io.EOF
				return io.EOF
			}

			if !isSilent(query) {
				l.Activity.Add(1)
			}

			l.queries <- &Query{
				link:  l,
				Query: query,
			}

			if !isSilent(query) {
				l.Activity.Done()
			}
		}
	}
}

func isSilent(q *link.Query) bool {
	return strings.HasPrefix(q.String(), ".")
}
