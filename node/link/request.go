package link

import (
	"github.com/cryptopunkscc/astrald/node/auth/id"
	"github.com/cryptopunkscc/astrald/node/mux"
	"io"
)

type Request struct {
	caller         id.Identity
	port           string
	localStream    mux.Stream
	remoteStreamID mux.StreamID
	link           *Link
}

// Accept the request
func (req *Request) Accept() (io.ReadWriteCloser, error) {
	err := req.sendAccept()
	if err != nil {
		_ = req.localStream.Close()
		return nil, err
	}

	return newConn(req.link, req.localStream, req.remoteStreamID), nil
}

// Reject the request
func (req Request) Reject() error {
	defer req.localStream.Close()
	return req.sendReject()
}

// Caller returns the auth.Identity of the caller
func (req Request) Caller() id.Identity {
	return req.caller
}

// Port returns the requested port
func (req Request) Port() string {
	return req.port
}

// sendAccept sends an accept frame to the caller
func (req *Request) sendAccept() error {
	return req.link.sendAccept(req.remoteStreamID, req.localStream.ID())
}

// sendReject sends a reject frame to the caller
func (req Request) sendReject() error {
	return req.link.sendReject(req.remoteStreamID)
}
