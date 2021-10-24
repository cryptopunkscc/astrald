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
)

// A Link is a multiplexed, authenticated connection between two parties allowing them to establish multiple data
// streams over a single secure channel.
type Link struct {
	requests chan Request
	conn     auth.Conn
	mux      *mux.Mux
	demux    *mux.StreamDemux
	closeCh  chan struct{}
}

const controlStreamID = 0

// New instantiates a new Link over the provided authenticated connection.
func New(conn auth.Conn) *Link {
	link := &Link{
		conn:     conn,
		mux:      mux.NewMux(conn),
		demux:    mux.NewStreamDemux(conn),
		requests: make(chan Request),
		closeCh:  make(chan struct{}),
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

	return newConn(inputStream, mux.NewOutputStream(link.mux, int(remoteStreamID))), nil
}

// Requests returns a channel to which incoming requests will be sent
func (link *Link) Requests() <-chan Request {
	return link.requests
}

// RemoteIdentity returns the auth.Identity of the remote party
func (link *Link) RemoteIdentity() id.Identity {
	return link.conn.RemoteIdentity()
}

// LocalIdentity returns the auth.Identity of the remote party
func (link *Link) LocalIdentity() id.Identity {
	return link.conn.LocalIdentity()
}

// RemoteAddr returns the network address of the remote party
func (link *Link) RemoteAddr() infra.Addr {
	return link.conn.RemoteAddr()
}

// Outbound returns true if we are the active party, false otherwise
func (link *Link) Outbound() bool {
	return link.conn.Outbound()
}

// WaitClose returns a channel which will be closed when the link is closed
func (link *Link) WaitClose() <-chan struct{} {
	return link.closeCh
}

// Close closes the link
func (link *Link) Close() error {
	return link.conn.Close()
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
	}
}
