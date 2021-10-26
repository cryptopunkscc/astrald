package link

import (
	mux "github.com/cryptopunkscc/astrald/mux"
	"sync"
	"time"
)

// Conn represents an open connection to the remote party's port. Shouldn't be instantiated directly.
type Conn struct {
	link         *Link
	inputStream  *mux.InputStream
	outputStream *mux.OutputStream
	query        string
	outbound     bool
	bytesRead    int
	bytesWritten int
	lastActive   time.Time
	closeCh      chan struct{}
	closed       bool
	mu           sync.Mutex
}

// newConn instantiates a new Conn and starts the necessary routines
func newConn(link *Link, query string, inputStream *mux.InputStream, outputStream *mux.OutputStream, outbound bool) *Conn {
	c := &Conn{
		link:         link,
		query:        query,
		closeCh:      make(chan struct{}),
		inputStream:  inputStream,
		outputStream: outputStream,
		outbound:     outbound,
		lastActive:   time.Now(),
	}

	if c.link != nil {
		c.link.addConn(c)
	}

	return c
}

func (conn *Conn) Read(p []byte) (n int, err error) {
	n, err = conn.inputStream.Read(p)
	if err != nil {
		conn.Close()
	}
	conn.addBytesRead(n)
	return n, err
}

func (conn *Conn) Write(p []byte) (n int, err error) {
	n, err = conn.outputStream.Write(p)
	if err == nil {
		conn.addBytesWritten(n)
	}
	return n, err
}

// Close closes the connection
func (conn *Conn) Close() error {
	conn.mu.Lock()
	defer conn.mu.Unlock()

	if conn.closed {
		return ErrStreamClosed
	}

	defer close(conn.closeCh)

	err := conn.outputStream.Close()
	if err != nil {
		return err
	}

	conn.closed = true
	return nil
}

func (conn *Conn) WaitClose() <-chan struct{} {
	return conn.closeCh
}

func (conn *Conn) Query() string {
	return conn.query
}

func (conn *Conn) BytesRead() int {
	return conn.bytesRead
}

func (conn *Conn) BytesWritten() int {
	return conn.bytesWritten
}

func (conn *Conn) Outbound() bool {
	return conn.outbound
}

func (conn *Conn) Idle() time.Duration {
	return time.Now().Sub(conn.lastActive)
}

func (conn *Conn) addBytesRead(n int) {
	conn.bytesRead += n
	conn.lastActive = time.Now()
	if conn.link != nil {
		conn.link.addBytesRead(n)
	}
}

func (conn *Conn) addBytesWritten(n int) {
	conn.bytesWritten += n
	conn.touch()
	if conn.link != nil {
		conn.link.addBytesWritten(n)
	}
}

func (conn *Conn) touch() {
	conn.lastActive = time.Now()
}
