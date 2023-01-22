package link

import (
	"context"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/auth"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/mux"
	"io"
	"log"
	"sync"
	"time"
)

const queryTimeout = 30 * time.Second

// A Link is a multiplexed, authenticated connection between two parties allowing them to establish multiple
// connections over any transport.
type Link struct {
	queries   chan *Query
	conns     map[*Conn]struct{}
	connsMu   sync.Mutex
	transport auth.Conn
	mux       *mux.Mux
	demux     *mux.StreamDemux
	closeCh   chan struct{}
	ctl       *mux.OutputStream
}

// New instantiates a new Link over the provided authenticated connection.
func New(conn auth.Conn) *Link {
	link := &Link{
		queries:   make(chan *Query),
		conns:     make(map[*Conn]struct{}),
		transport: conn,
		mux:       mux.NewMux(conn),
		demux:     mux.NewStreamDemux(conn),
		closeCh:   make(chan struct{}),
	}

	link.ctl = link.mux.ControlStream()
	go link.run()

	return link
}

// Query requests a connection to the remote party's port
func (link *Link) Query(ctx context.Context, query string) (*Conn, error) {
	// Reserve a local mux stream for the response
	inputStream, err := link.AllocInputStream()
	if err != nil {
		return nil, err
	}

	// Send a query
	err = cslq.Encode(link.ctl, "v", queryData{
		StreamID: inputStream.ID(),
		Query:    query,
	})
	if err != nil {
		inputStream.Discard()
		link.Close()
		return nil, err
	}

	var errCh = make(chan error, 1)
	var resCh = make(chan int, 1)

	go func() {
		var remoteStreamID int
		err := cslq.Decode(inputStream, "s", &remoteStreamID)
		switch {
		case err == nil:
			resCh <- remoteStreamID
		case errors.Is(err, io.EOF):
			errCh <- ErrRejected
		default:
			link.Close()
			errCh <- err
		}
		return
	}()

	var remoteStreamID int

	select {
	case err := <-errCh:
		return nil, err

	case <-ctx.Done():
		inputStream.Discard()
		return nil, ctx.Err()

	case <-time.After(queryTimeout):
		inputStream.Discard()
		link.Close()
		return nil, ErrIOTimeout

	case remoteStreamID = <-resCh:
	}

	outputStream := link.mux.Stream(remoteStreamID)
	conn := newConn(inputStream, outputStream, true, query)
	conn.Attach(link)

	return conn, nil
}

// AllocInputStream allocates and returns a new input stream on the multiplexer
func (link *Link) AllocInputStream() (*mux.InputStream, error) {
	return link.demux.AllocStream()
}

// OutputStream returns an output stream representing the provided stream id on the multiplexer
func (link *Link) OutputStream(id int) *mux.OutputStream {
	return link.mux.Stream(id)
}

// Queries returns a channel to which incoming queries will be sent
func (link *Link) Queries() <-chan *Query {
	return link.queries
}

// RemoteIdentity returns the auth.Identity of the remote party
func (link *Link) RemoteIdentity() id.Identity {
	return link.transport.RemoteIdentity()
}

// LocalIdentity returns the auth.Identity of the remote party
func (link *Link) LocalIdentity() id.Identity {
	return link.transport.LocalIdentity()
}

func (link *Link) LocalAddr() infra.Addr {
	return link.transport.LocalAddr()
}

// RemoteAddr returns the network address of the remote party
func (link *Link) RemoteAddr() infra.Addr {
	return link.transport.RemoteAddr()
}

func (link *Link) Network() string {
	if a := link.LocalAddr(); a != nil {
		return a.Network()
	}
	if a := link.RemoteAddr(); a != nil {
		return a.Network()
	}
	return "unknown"
}

// Outbound returns true if we are the active party, false otherwise
func (link *Link) Outbound() bool {
	return link.transport.Outbound()
}

// Wait returns a channel which will be closed when the link is closed
func (link *Link) Wait() <-chan struct{} {
	return link.closeCh
}

// Close closes the link
func (link *Link) Close() error {
	return link.transport.Close()
}

func (link *Link) Conns() <-chan *Conn {
	link.connsMu.Lock()
	defer link.connsMu.Unlock()

	ch := make(chan *Conn, len(link.conns))
	for c := range link.conns {
		ch <- c
	}
	close(ch)
	return ch
}

func (link *Link) remove(conn *Conn) error {
	link.connsMu.Lock()
	defer link.connsMu.Unlock()

	if _, found := link.conns[conn]; !found {
		return errors.New("conn not found")
	}

	delete(link.conns, conn)
	return nil
}

func (link *Link) add(conn *Conn) error {
	link.connsMu.Lock()
	defer link.connsMu.Unlock()

	if _, found := link.conns[conn]; found {
		return errors.New("already added")
	}

	link.conns[conn] = struct{}{}

	return nil
}

// sendQuery writes a frame containing the request to the control stream
func (link *Link) sendQuery(query string, streamID int) error {
	return cslq.Encode(link.ctl, "v", queryData{
		StreamID: streamID,
		Query:    query,
	})
}

func (link *Link) run() error {
	defer close(link.queries)
	defer close(link.closeCh)

	ctl, err := link.demux.ControlStream()
	if err != nil {
		return err
	}

	for {
		var q queryData

		// read next query
		err := cslq.Decode(ctl, "v", &q)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				log.Println("link.Link.run() error:", err)
			}
			return err
		}

		// process it
		err = link.processQuery(q)
		if err != nil {
			log.Println("link.Link.run() error:", err)
		}
	}
}

func (link *Link) processQuery(q queryData) error {
	var err error

	// Prepare request object
	query := &Query{
		link:  link,
		out:   link.mux.Stream(q.StreamID),
		query: q.Query,
	}

	// Reserve local channel
	query.in, err = link.demux.AllocStream()
	if err != nil {
		query.out.Close()
		return fmt.Errorf("cannot allocate input stream: %w", err)
	}

	link.queries <- query

	return nil
}
