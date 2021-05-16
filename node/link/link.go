package link

import (
	"encoding/binary"
	"github.com/cryptopunkscc/astrald/node/auth"
	"github.com/cryptopunkscc/astrald/node/link/proto"
	"github.com/cryptopunkscc/astrald/node/mux"
	"io"
	"log"
)

// A Link is a multiplexed, authenticated connection between two parties allowing them to establish multiple data
// streams over a single secure channel.
type Link struct {
	requests chan Request
	conn     auth.Conn
	mux      *mux.Mux
}

// New instantiates a new Link over the provided authenticated connection.
func New(conn auth.Conn) *Link {
	link := &Link{
		conn:     conn,
		mux:      mux.New(conn),
		requests: make(chan Request),
	}
	go link.process()
	return link
}

// Requests returns a channel to which incoming requests will be sent
func (link *Link) Requests() <-chan Request {
	return link.requests
}

// Open requests a connection to the remote party's port
func (link *Link) Open(port string) (io.ReadWriteCloser, error) {
	// Reserve a local mux stream
	localStream, err := link.mux.Stream()
	if err != nil {
		return nil, err
	}

	// Send a request
	err = link.sendRequest(port, localStream)
	if err != nil {
		_ = localStream.Close()
		return nil, err
	}

	// Wait for response (the first frame received in the local channel)
	// TODO: handle some kind of timeout?
	response := <-localStream.Frames()

	// An accept message is always 2 bytes long, everything else is a reject
	if len(response.Data) != 2 {
		_ = localStream.Close()
		return nil, ErrRejected
	}

	// Parse the accept message
	remoteStreamID := mux.StreamID(binary.BigEndian.Uint16(response.Data[0:2]))

	return newConn(link, localStream, remoteStreamID), nil
}

// RemoteIdentity returns the auth.Identity of the remote party
func (link *Link) RemoteIdentity() auth.Identity {
	return link.conn.RemoteIdentity()
}

// Outbound returns true if we are the active party, false otherwise
func (link *Link) Outbound() bool {
	return link.conn.Outbound()
}

// sendRequest writes a frame containing the request to the control stream
func (link *Link) sendRequest(port string, localStream mux.Stream) error {
	return link.mux.Write(mux.Frame{
		StreamID: mux.ControlStreamID,
		Data: proto.Request{
			StreamID: uint16(localStream.ID()),
			Port:     port,
		}.Bytes(),
	})
}

// process incoming control frames
func (link *Link) process() {
	log.Println("link established:", link.RemoteIdentity())
	for ctlFrame := range link.mux.Control() {
		// Parse the request frame
		rawRequest, err := proto.ParseRequest(ctlFrame.Data)
		if err != nil {
			// TODO: a port in closing state should ignore all frame until a RST (0-len payload) frame is received
			log.Println("error parsing control frame:", err)
			continue // skip to the next control frame
		}

		request := Request{
			caller:         link.RemoteIdentity(),
			remoteStreamID: mux.StreamID(rawRequest.StreamID),
			port:           rawRequest.Port,
			link:           link,
		}

		// Reserve local channel
		request.localStream, err = link.mux.Stream()
		if err != nil {
			_ = link.sendReject(mux.StreamID(rawRequest.StreamID))
			continue
		}

		link.requests <- request
	}

	//TODO: proper logging
	log.Println("link lost:", link.RemoteIdentity())
}

// sendAccept sends an accept frame to the remote stream
func (link *Link) sendAccept(remoteStreamID mux.StreamID, localStreamID mux.StreamID) error {
	data := make([]byte, 2)
	binary.BigEndian.PutUint16(data[:], uint16(localStreamID))
	return link.mux.Write(mux.Frame{
		StreamID: remoteStreamID,
		Data:     data,
	})
}

// sendReject sends a reject frame (0-length) to the remote stream
func (link *Link) sendReject(remoteStreamID mux.StreamID) error {
	return link.mux.Write(mux.Frame{
		StreamID: remoteStreamID,
		Data:     nil,
	})
}
