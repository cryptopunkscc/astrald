package link

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/auth"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/link/proto"
	"github.com/cryptopunkscc/astrald/mux"
	"io"
	"log"
	"sync"
)

// A Link is a multiplexed, authenticated connection between two parties allowing them to establish multiple
// connections over any transport.
type Link struct {
	queries   chan *Query
	conns     map[*Conn]struct{}
	connsMu   sync.Mutex
	transport auth.Conn
	mux       *mux.Mux
	demux     *mux.StreamDemux
	closed    chan struct{}
}

const controlStreamID = 0

// New instantiates a new Link over the provided authenticated connection.
func New(conn auth.Conn) *Link {
	link := &Link{
		queries:   make(chan *Query),
		conns:     make(map[*Conn]struct{}),
		transport: conn,
		mux:       mux.NewMux(conn),
		demux:     mux.NewStreamDemux(conn),
		closed:    make(chan struct{}),
	}
	go func() {
		defer close(link.closed)

		err := link.processQueries()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				log.Println("closing link:", err)
			}
		}
	}()
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
	err = link.sendQuery(query, inputStream.ID())
	if err != nil {
		inputStream.Close()
		return nil, err
	}

	remoteBytes := make([]byte, 2)

	_, err = inputStream.Read(remoteBytes)

	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, ErrRejected
		}
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	// Parse accept message
	remoteStreamID := binary.BigEndian.Uint16(remoteBytes[0:2])

	outputStream := link.mux.Stream(int(remoteStreamID))

	conn := NewConn(inputStream, outputStream, true, query)
	conn.Attach(link)

	return conn, nil
}

// AllocInputStream allocates and returns a new input stream on the multiplexer
func (link *Link) AllocInputStream() (*mux.InputStream, error) {
	return link.demux.Stream()
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
	return link.closed
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
	return link.mux.Write(controlStreamID, proto.MakeQuery(streamID, query))
}

func (link *Link) processQueries() error {
	buf := make([]byte, mux.MaxPayload)
	ctlStream := link.demux.DefaultStream()

	defer close(link.queries)
	for {
		n, err := ctlStream.Read(buf)
		if err != nil {
			return fmt.Errorf("control stream error: %w", err)
		}

		err = link.processQueryData(buf[:n])
		if err != nil {
			log.Println("error processing query:", err)
		}
	}
}

func (link *Link) processQueryData(buf []byte) error {
	// Parse control message
	remoteStreamID, queryString, err := proto.ParseQuery(buf)
	if err != nil {
		return fmt.Errorf("parse query: %w", err)
	}

	// Prepare request object
	outputStream := link.mux.Stream(remoteStreamID)
	query := &Query{
		link:         link,
		outputStream: outputStream,
		query:        queryString,
	}

	// Reserve local channel
	query.inputStream, err = link.demux.Stream()
	if err != nil {
		_ = outputStream.Close()
		return fmt.Errorf("cannot allocate input stream: %w", err)
	}

	link.queries <- query
	return nil
}
