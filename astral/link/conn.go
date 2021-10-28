package link

import (
	"io"
	"sync"
	"time"
)

// Conn represents an open connection to the remote party's port. Shouldn't be instantiated directly.
type Conn struct {
	inputStream  io.Reader
	outputStream io.WriteCloser
	query        string
	outbound     bool
	bytesRead    int
	bytesWritten int
	lastActive   time.Time
	closeCh      chan struct{}
	closed       bool
	mu           sync.Mutex
	container    Toucher
}

// newConn instantiates a new Conn and starts the necessary routines
func newConn(outer Toucher, inputStream io.Reader, outputStream io.WriteCloser, outbound bool, query string) *Conn {
	c := &Conn{
		container:    outer,
		query:        query,
		closeCh:      make(chan struct{}),
		inputStream:  inputStream,
		outputStream: outputStream,
		outbound:     outbound,
		lastActive:   time.Now(),
	}

	return c
}

func (conn *Conn) Read(p []byte) (n int, err error) {
	n, err = conn.inputStream.Read(p)
	conn.addBytesRead(n)
	if err != nil {
		_ = conn.Close()
	}
	conn.Touch()
	return n, err
}

func (conn *Conn) Write(p []byte) (n int, err error) {
	n, err = conn.outputStream.Write(p)
	conn.addBytesWritten(n)
	conn.Touch()
	return n, err
}

// Close closes the connection
func (conn *Conn) Close() error {
	conn.mu.Lock()
	defer conn.mu.Unlock()

	if conn.closed {
		return ErrStreamClosed
	}
	conn.closed = true

	defer close(conn.closeCh)

	err := conn.outputStream.Close()
	if err != nil {
		return err
	}

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
}

func (conn *Conn) addBytesWritten(n int) {
	conn.bytesWritten += n
}

func (conn *Conn) Touch() {
	conn.lastActive = time.Now()
	if conn.container != nil {
		conn.container.Touch()
	}
}
