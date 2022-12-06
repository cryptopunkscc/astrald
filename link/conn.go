package link

import (
	"github.com/cryptopunkscc/astrald/mux"
	"sync"
)

// Conn represents an open connection to the remote party's port. Shouldn't be instantiated directly.
type Conn struct {
	in         *mux.InputStream
	out        *mux.OutputStream
	fallbackIn *mux.InputStream
	query      string
	outbound   bool
	closeOnce  sync.Once
	closeCh    chan struct{}
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

	if err != nil {
		if conn.fallbackIn != nil {
			conn.in = conn.fallbackIn
			conn.fallbackIn = nil
			return conn.Read(p)
		}
		conn.Close()
	}

	return
}

func (conn *Conn) Write(p []byte) (n int, err error) {
	conn.outMu.Lock()
	defer conn.outMu.Unlock()

	return conn.out.Write(p)
}

// Close closes the connection
func (conn *Conn) Close() (err error) {
	conn.outMu.Lock()
	defer conn.outMu.Unlock()

	err = ErrAlreadyClosed
	conn.closeOnce.Do(func() {
		err = conn.out.Close()
		if err != nil {
			conn.in.Close()
		}
		conn.Attach(nil)
		close(conn.closeCh)
	})

	return
}

func (conn *Conn) InputStream() *mux.InputStream {
	return conn.in
}

func (conn *Conn) OutputStream() *mux.OutputStream {
	return conn.out
}

func (conn *Conn) SetFallbackInputStream(fallbackInputStream *mux.InputStream) {
	conn.fallbackIn = fallbackInputStream
}

func (conn *Conn) ReplaceOutputStream(stream *mux.OutputStream) {
	conn.outMu.Lock()
	defer conn.outMu.Unlock()

	conn.out.Close()
	conn.out = stream
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
