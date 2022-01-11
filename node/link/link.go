package link

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/link"
	"github.com/cryptopunkscc/astrald/sig"
	"log"
	"strings"
	"sync"
	"time"
)

const idlePingInterval = 15 * time.Second
const activePingInterval = 1 * time.Second
const pingTimeout = 15 * time.Second

var _ sig.Idler = &Link{}

// Link wraps an astral.Link and adds activity tracking
type Link struct {
	sig.Activity
	*link.Link
	queries       chan *Query
	events        chan Event
	conns         map[*Conn]struct{}
	mu            sync.Mutex
	establishedAt time.Time
	roundtrip     time.Duration
}

func Wrap(link *link.Link) *Link {
	l := &Link{
		Link:          link,
		queries:       make(chan *Query),
		events:        make(chan Event),
		conns:         make(map[*Conn]struct{}, 0),
		establishedAt: time.Now(),
		roundtrip:     999 * time.Second, // assume super slow before first ping
	}
	l.Touch()
	go l.handleQueries()
	go l.monitorPing()
	return l
}

func (link *Link) Query(ctx context.Context, query string) (*Conn, error) {
	if len(query) == 0 {
		return nil, errors.New("empty query")
	}

	// silent queries do not affect activity
	if !(query[0] == '.') {
		link.Activity.Add(1)
		defer link.Activity.Done()
	}

	linkConn, err := link.Link.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	conn := wrapConn(linkConn)
	conn.Attach(link)

	link.events <- EventConnEstablished{conn}

	return conn, nil
}

// RoundTrip returns link's round trip time
func (link *Link) RoundTrip() time.Duration {
	return link.roundtrip
}

func (link *Link) Queries() <-chan *Query {
	return link.queries
}

func (link *Link) Events() <-chan Event {
	return link.events
}

func (link *Link) Conns() <-chan *Conn {
	link.mu.Lock()
	defer link.mu.Unlock()

	ch := make(chan *Conn, len(link.conns))
	for link, _ := range link.conns {
		ch <- link
	}
	close(ch)
	return ch
}

func (link *Link) add(conn *Conn) {
	link.mu.Lock()
	defer link.mu.Unlock()

	// skip duplicates
	if _, found := link.conns[conn]; found {
		return
	}

	link.conns[conn] = struct{}{}
	link.Activity.Add(1)
}

func (link *Link) monitorPing() {
	for {
		if err := link.ping(); err != nil {
			log.Println("ping error:", err)
			link.Close()
		}

		// if the connection is active, we want to monitor ping more often
		i := idlePingInterval
		if link.Activity.Idle() == 0 {
			i = activePingInterval
		}

		select {
		case <-time.After(i):
		case <-link.Wait():
			return
		}
	}
}

func (link *Link) remove(conn *Conn) error {
	link.mu.Lock()
	defer link.mu.Unlock()

	if _, found := link.conns[conn]; !found {
		return errors.New("not found")
	}
	delete(link.conns, conn)
	link.Activity.Done()

	return nil
}

func (link *Link) handleQueries() {
	defer close(link.queries)
	for query := range link.Link.Queries() {
		if !isSilent(query) {
			link.Activity.Add(1)
		}

		link.queries <- &Query{
			link:  link,
			Query: query,
		}

		if !isSilent(query) {
			link.Activity.Done()
		}
	}
}

func (link *Link) ping() error {
	pingCh := make(chan time.Duration, 1)

	go func() {
		t0 := time.Now()
		ctx, _ := context.WithTimeout(context.Background(), pingTimeout)
		conn, err := link.Query(ctx, ".ping")

		if errors.Is(err, context.Canceled) {
			return
		}

		pingCh <- time.Now().Sub(t0)

		if err == nil {
			conn.Close()
		}
	}()

	select {
	case d := <-pingCh:
		link.roundtrip = d
	case <-time.After(pingTimeout):
		link.Close()
		return errors.New("ping timeout")
	}

	return nil
}

func isSilent(q *link.Query) bool {
	return strings.HasPrefix(q.String(), ".")
}
