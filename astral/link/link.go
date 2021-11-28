package link

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/astral/link/proto"
	"github.com/cryptopunkscc/astrald/auth"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/mux"
	"github.com/cryptopunkscc/astrald/sig"
	"io"
	"log"
	"sync"
	"time"
)

var _ sig.Idler = &Link{}

// A Link is a multiplexed, authenticated connection between two parties allowing them to establish multiple data
// streams over a single secure channel.
type Link struct {
	activity  sig.Activity
	requests  chan Request
	conns     []*Conn
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
		requests:  make(chan Request),
		conns:     make([]*Conn, 0),
		transport: conn,
		mux:       mux.NewMux(conn),
		demux:     mux.NewStreamDemux(conn),
		closed:    make(chan struct{}),
	}
	link.activity.Touch()
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
func (link *Link) Query(query string) (*Conn, error) {
	link.activity.Add(1)
	defer link.activity.Done()

	// Reserve a local mux stream
	inputStream, err := link.demux.Stream()
	if err != nil {
		return nil, err
	}

	// Send a request
	err = link.sendQuery(query, inputStream.StreamID())
	if err != nil {
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

	outputStream := mux.NewOutputStream(link.mux, int(remoteStreamID))

	return link.addConn(inputStream, outputStream, true, query), nil
}

// Requests returns a channel to which incoming requests will be sent
func (link *Link) Requests() <-chan Request {
	return link.requests
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

func (link *Link) Idle() time.Duration {
	return link.activity.Idle()
}

// Close closes the link
func (link *Link) Close() error {
	return link.transport.Close()
}

func (link *Link) Conns() <-chan *Conn {
	link.connsMu.Lock()
	defer link.connsMu.Unlock()

	ch := make(chan *Conn, len(link.conns))
	for _, c := range link.conns {
		ch <- c
	}
	close(ch)
	return ch
}

// sendQuery writes a frame containing the request to the control stream
func (link *Link) sendQuery(query string, streamID int) error {
	return link.mux.Write(controlStreamID, proto.MakeQuery(streamID, query))
}

func (link *Link) processQueries() error {
	buf := make([]byte, mux.MaxPayload)
	ctlStream := link.demux.DefaultStream()

	defer close(link.requests)
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
	link.activity.Add(1)
	defer link.activity.Done()

	// Parse control message
	remoteStreamID, query, err := proto.ParseQuery(buf)
	if err != nil {
		return fmt.Errorf("parse query: %w", err)
	}

	// Prepare request object
	outputStream := link.mux.Stream(remoteStreamID)
	request := Request{
		link:         link,
		outputStream: outputStream,
		query:        query,
	}

	// Reserve local channel
	request.inputStream, err = link.demux.Stream()
	if err != nil {
		_ = outputStream.Close()
		return fmt.Errorf("cannot allocate input stream: %w", err)
	}

	link.requests <- request
	return nil
}

func (link *Link) addConn(inputStream io.Reader, outputStream io.WriteCloser, outbound bool, query string) *Conn {
	link.connsMu.Lock()
	defer link.connsMu.Unlock()

	conn := newConn(inputStream, outputStream, outbound, query)
	link.conns = append(link.conns, conn)
	link.activity.Add(1)

	go func() {
		<-conn.Wait()
		link.removeConn(conn)
	}()

	return conn
}

func (link *Link) removeConn(conn *Conn) {
	link.connsMu.Lock()
	defer link.connsMu.Unlock()

	for i, c := range link.conns {
		if c == conn {
			link.conns = append(link.conns[:i], link.conns[i+1:]...)
			link.activity.Done()
			return
		}
	}
}
