package link

import (
	"encoding/binary"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mux"
	"io"
)

type Request struct {
	caller       id.Identity
	query        string
	inputStream  *mux.InputStream
	outputStream *mux.OutputStream
}

// Accept the request
func (req *Request) Accept() (io.ReadWriteCloser, error) {
	err := binary.Write(req.outputStream, binary.BigEndian, uint16(req.inputStream.StreamID()))
	if err != nil {
		return nil, err
	}

	return newConn(req.inputStream, req.outputStream), nil
}

// Reject the request
func (req Request) Reject() error {
	req.inputStream.Close()

	return req.outputStream.Close()
}

// Caller returns the auth.Identity of the caller
func (req Request) Caller() id.Identity {
	return req.caller
}

// Query returns the requested port
func (req Request) Query() string {
	return req.query
}
