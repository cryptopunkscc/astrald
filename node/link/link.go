package link

import (
	"errors"
	"github.com/cryptopunkscc/astrald/astral/link"
	"github.com/cryptopunkscc/astrald/auth"
	"github.com/cryptopunkscc/astrald/sig"
	"sync"
)

var _ sig.Idler = &Link{}

// Link wraps an astral.Link and adds activity tracking
type Link struct {
	sig.Activity
	*link.Link
	queries chan *Query
	conns   map[*Conn]struct{}
	mu      sync.Mutex
}

func New(conn auth.Conn) *Link {
	return Wrap(link.New(conn))
}

func Wrap(link *link.Link) *Link {
	l := &Link{
		Link:    link,
		queries: make(chan *Query),
		conns:   make(map[*Conn]struct{}, 0),
	}
	l.Touch()
	go l.handleQueries()
	return l
}

func (link *Link) Query(query string) (*Conn, error) {
	link.Activity.Add(1)
	defer link.Activity.Done()

	linkConn, err := link.Link.Query(query)
	if err != nil {
		return nil, err
	}

	conn := wrapConn(linkConn)
	link.add(conn)

	return conn, err
}

func (link *Link) Queries() <-chan *Query {
	return link.queries
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

	go func() {
		link.Activity.Add(1)
		defer link.Activity.Done()

		// remove the connection after it closes
		<-conn.Wait()
		link.remove(conn)
	}()
}

func (link *Link) remove(conn *Conn) error {
	link.mu.Lock()
	defer link.mu.Unlock()

	if _, found := link.conns[conn]; !found {
		return errors.New("not found")
	}
	delete(link.conns, conn)

	return nil
}

func (link *Link) handleQueries() {
	defer close(link.queries)
	for r := range link.Link.Queries() {
		link.Activity.Add(1)
		link.queries <- &Query{
			link:  link,
			Query: r,
		}
		link.Activity.Done()
	}
}
