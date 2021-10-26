package link

import (
	"encoding/binary"
	"errors"
	"github.com/cryptopunkscc/astrald/astral/link/proto"
	"github.com/cryptopunkscc/astrald/auth"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/mux"
	"log"
	"sync"
	"time"
)

// A Link is a multiplexed, authenticated connection between two parties allowing them to establish multiple data
// streams over a single secure channel.
type Link struct {
	requests     chan Request
	conns        []*Conn
	connsMu      sync.Mutex
	transport    auth.Conn
	mux          *mux.Mux
	demux        *mux.StreamDemux
	closeCh      chan struct{}
	bytesRead    int
	bytesWritten int
	lastActive   time.Time
}

const controlStreamID = 0

// New instantiates a new Link over the provided authenticated connection.
func New(conn auth.Conn) *Link {
	link := &Link{
		requests:   make(chan Request),
		conns:      make([]*Conn, 0),
		transport:  conn,
		mux:        mux.NewMux(conn),
		demux:      mux.NewStreamDemux(conn),
		closeCh:    make(chan struct{}),
		lastActive: time.Now(),
	}
	go func() {
		_ = link.handleControl()
		close(link.requests)
		close(link.closeCh)
	}()
	return link
}

// Query requests a connection to the remote party's port
func (link *Link) Query(query string) (*Conn, error) {
	defer link.touch()

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
		return nil, ErrRejected
	}

	// Parse the accept message
	remoteStreamID := binary.BigEndian.Uint16(remoteBytes[0:2])

	return newConn(link, query, inputStream, mux.NewOutputStream(link.mux, int(remoteStreamID)), true), nil
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

// RemoteAddr returns the network address of the remote party
func (link *Link) RemoteAddr() infra.Addr {
	return link.transport.RemoteAddr()
}

// Outbound returns true if we are the active party, false otherwise
func (link *Link) Outbound() bool {
	return link.transport.Outbound()
}

// WaitClose returns a channel which will be closed when the link is closed
func (link *Link) WaitClose() <-chan struct{} {
	return link.closeCh
}

// Close closes the link
func (link *Link) Close() error {
	return link.transport.Close()
}

func (link *Link) Connections() <-chan *Conn {
	link.connsMu.Lock()
	defer link.connsMu.Unlock()

	ch := make(chan *Conn, len(link.conns))
	for _, c := range link.conns {
		ch <- c
	}
	close(ch)
	return ch
}

func (link *Link) BytesRead() int {
	return link.bytesRead
}

func (link *Link) BytesWritten() int {
	return link.bytesWritten
}

func (link *Link) Idle() time.Duration {
	return time.Now().Sub(link.lastActive)
}

// sendQuery writes a frame containing the request to the control stream
func (link *Link) sendQuery(query string, streamID int) error {
	return link.mux.Write(controlStreamID, proto.MakeQuery(streamID, query))
}

func (link *Link) handleControl() error {
	buf := make([]byte, mux.MaxPayload)
	ctlStream := link.demux.DefaultStream()

	for {
		n, err := ctlStream.Read(buf)
		if err != nil {
			return errors.New("control stream error")
		}

		// Parse control message
		remoteStreamID, query, err := proto.ParseQuery(buf[:n])
		if err != nil {
			return errors.New("control frame parse error")
		}

		// Prepare request object
		outputStream := link.mux.Stream(remoteStreamID)
		request := Request{
			caller:       link.RemoteIdentity(),
			outputStream: outputStream,
			query:        query,
		}

		// Reserve local channel
		request.inputStream, err = link.demux.Stream()
		if err != nil {
			log.Println("error allocating new input stream:", err)
			outputStream.Close()
			continue
		}

		link.requests <- request

		link.touch()
	}
}

func (link *Link) addConn(conn *Conn) {
	link.connsMu.Lock()
	defer link.connsMu.Unlock()

	link.conns = append(link.conns, conn)
	link.touch()

	go func() {
		<-conn.WaitClose()
		link.removeConn(conn)
	}()
}

func (link *Link) removeConn(conn *Conn) {
	link.connsMu.Lock()
	defer link.connsMu.Unlock()

	for i, c := range link.conns {
		if c == conn {
			link.conns = append(link.conns[:i], link.conns[i+1:]...)
			link.touch()
			return
		}
	}
}

func (link *Link) addBytesRead(n int) {
	link.bytesRead += n
	link.touch()
}

func (link *Link) addBytesWritten(n int) {
	link.bytesWritten += n
	link.touch()
}

func (link *Link) touch() {
	link.lastActive = time.Now()
}
