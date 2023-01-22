package link

import (
	"github.com/cryptopunkscc/astrald/mux"
	"sync"
	"sync/atomic"
)

// Conn represents an open connection to the remote party's port. Shouldn't be instantiated directly.
type Conn struct {
	in         *mux.InputStream
	out        *mux.OutputStream
	fallbackIn *mux.InputStream
	fallbackMu sync.Mutex
	query      string
	outbound   bool
	closeCh    chan struct{}
	closed     atomic.Bool
	outMu      sync.Mutex
	link       *Link
	linkMu     sync.Mutex
}

// newConn instantiates a new Conn
func newConn(inputStream *mux.InputStream, outputStream *mux.OutputStream, outbound bool, query string) *Conn {
	return &Conn{
		query:    query,
		closeCh:  make(chan struct{}),
		in:       inputStream,
		out:      outputStream,
		outbound: outbound,
	}
}

func (conn *Conn) Read(p []byte) (n int, err error) {
	n, err = conn.in.Read(p)

	if n > 0 {
		return
	}
	if err == nil {
		return
	}

	// use fallback input if set
	conn.fallbackMu.Lock()
	if conn.fallbackIn != nil {
		conn.in = conn.fallbackIn
		conn.fallbackIn = nil
		conn.fallbackMu.Unlock()
		return conn.Read(p)
	}
	conn.fallbackMu.Unlock()

	conn.Close()
	return
}

func (conn *Conn) Write(p []byte) (n int, err error) {
	conn.outMu.Lock()
	defer conn.outMu.Unlock()

	if conn.closed.Load() {
		return 0, ErrClosed
	}

	return conn.out.Write(p)
}

// Close closes the connection
func (conn *Conn) Close() error {
	if conn.closed.CompareAndSwap(false, true) {
		conn.out.Close()
		conn.Attach(nil)
		close(conn.closeCh)
	}
	return nil
}

func (conn *Conn) InputStream() *mux.InputStream {
	return conn.in
}

func (conn *Conn) OutputStream() *mux.OutputStream {
	return conn.out
}

func (conn *Conn) SetFallbackInputStream(fallback *mux.InputStream) {
	conn.fallbackMu.Lock()
	defer conn.fallbackMu.Unlock()
	conn.fallbackIn = fallback
}

func (conn *Conn) ReplaceOutputStream(replacement *mux.OutputStream) error {
	conn.outMu.Lock()
	defer conn.outMu.Unlock()

	if conn.closed.Load() {
		return ErrClosed
	}

	conn.out.Close()
	conn.out = replacement
	return nil
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
	}

	conn.link = link

	if conn.link != nil {
		conn.link.add(conn)
	}
}
