package link

import (
	"github.com/cryptopunkscc/astrald/node/mux"
	"sync"
)

// Conn represents an open connection to the remote party's port. Shouldn't be instantiated directly.
type Conn struct {
	inputStream  *mux.InputStream
	outputStream *mux.OutputStream
	mu           sync.Mutex
	closed       bool
}

// newConn instantiates a new Conn and starts the necessary routines
func newConn(inputStream *mux.InputStream, outputStream *mux.OutputStream) *Conn {
	return &Conn{
		inputStream:  inputStream,
		outputStream: outputStream,
	}
}

func (conn *Conn) Read(p []byte) (n int, err error) {
	read, err := conn.inputStream.Read(p)
	if err != nil {
		conn.Close()
	}
	return read, err
}

func (conn *Conn) Write(p []byte) (n int, err error) {
	return conn.outputStream.Write(p)
}

// Close closes the connection
func (conn *Conn) Close() error {
	conn.mu.Lock()
	defer conn.mu.Unlock()

	if conn.closed {
		return ErrStreamClosed
	}

	err := conn.outputStream.Close()
	if err != nil {
		return err
	}

	conn.closed = true
	return nil
}
