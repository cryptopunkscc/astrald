package link

import (
	"github.com/cryptopunkscc/astrald/mux"
	"sync"
)

// Conn represents an open connection to the remote party's port. Shouldn't be instantiated directly.
type Conn struct {
	inputStream         *mux.InputStream
	outputStream        *mux.OutputStream
	fallbackInputStream *mux.InputStream
	query               string
	outbound            bool
	closeCh             chan struct{}
	closed              bool
	outputMu            sync.Mutex
	link                *Link
	linkMu              sync.Mutex
}

// NewConn instantiates a new Conn
func NewConn(inputStream *mux.InputStream, outputStream *mux.OutputStream, outbound bool, query string) *Conn {
	c := &Conn{
		query:        query,
		closeCh:      make(chan struct{}),
		inputStream:  inputStream,
		outputStream: outputStream,
		outbound:     outbound,
	}

	go func() {
		<-c.Wait()
		c.Attach(nil)
	}()

	return c
}

func (conn *Conn) Read(p []byte) (n int, err error) {
	n, err = conn.inputStream.Read(p)

	if err != nil {
		if conn.fallbackInputStream != nil {
			conn.inputStream = conn.fallbackInputStream
			conn.fallbackInputStream = nil
			return conn.Read(p)
		}
		_ = conn.Close()
	}
	return n, err
}

func (conn *Conn) Write(p []byte) (n int, err error) {
	conn.outputMu.Lock()
	defer conn.outputMu.Unlock()

	n, err = conn.outputStream.Write(p)
	return n, err
}

// Close closes the connection
func (conn *Conn) Close() error {
	conn.outputMu.Lock()
	defer conn.outputMu.Unlock()

	if conn.closed {
		return ErrAlreadyClosed
	}
	conn.closed = true

	defer close(conn.closeCh)

	err := conn.outputStream.Close()
	if err != nil {
		return err
	}

	return nil
}

func (conn *Conn) InputStream() *mux.InputStream {
	return conn.inputStream
}

func (conn *Conn) OutputStream() *mux.OutputStream {
	return conn.outputStream
}

func (conn *Conn) SetFallbackInputStream(fallbackInputStream *mux.InputStream) {
	conn.fallbackInputStream = fallbackInputStream
}

func (conn *Conn) ReplaceOutputStream(stream *mux.OutputStream) {
	conn.outputMu.Lock()
	defer conn.outputMu.Unlock()

	conn.outputStream.Close()
	conn.outputStream = stream
}

// Wait returns a channel that will be closed when the connection closes
func (conn *Conn) Wait() <-chan struct{} {
	return conn.closeCh
}

// Query returns the query string used to establish the connection
func (conn *Conn) Query() string {
	return conn.query
}

// Outbound returns true if connection is outbound
func (conn *Conn) Outbound() bool {
	return conn.outbound
}

// Attach attaches this connection to a link, attach to nil to detach
func (conn *Conn) Attach(link *Link) {
	conn.linkMu.Lock()
	defer conn.linkMu.Unlock()

	if conn.link != nil {
		conn.link.remove(conn)
		conn.link = nil
	}

	if link != nil {
		conn.link = link
		conn.link.add(conn)
	}
}
