package link

import (
	async "github.com/cryptopunkscc/astrald/sync"
	"io"
	"sync"
	"time"
)

var _ async.Idler = &Conn{}

// Conn represents an open connection to the remote party's port. Shouldn't be instantiated directly.
type Conn struct {
	activity     async.Activity
	inputStream  io.Reader
	outputStream io.WriteCloser
	query        string
	outbound     bool
	closeCh      chan struct{}
	closed       bool
	mu           sync.Mutex
}

// newConn instantiates a new Conn and starts the necessary routines
func newConn(inputStream io.Reader, outputStream io.WriteCloser, outbound bool, query string) *Conn {
	c := &Conn{
		query:        query,
		closeCh:      make(chan struct{}),
		inputStream:  inputStream,
		outputStream: outputStream,
		outbound:     outbound,
	}

	c.activity.Touch()

	return c
}

func (conn *Conn) Read(p []byte) (n int, err error) {
	defer conn.activity.Touch()

	n, err = conn.inputStream.Read(p)
	//conn.AddBytesRead(n)
	if err != nil {
		_ = conn.Close()
	}
	return n, err
}

func (conn *Conn) Write(p []byte) (n int, err error) {
	defer conn.activity.Touch()

	n, err = conn.outputStream.Write(p)
	//conn.AddBytesWritten(n)
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

func (conn *Conn) Wait() <-chan struct{} {
	return conn.closeCh
}

func (conn *Conn) Idle() time.Duration {
	return conn.activity.Idle()
}

func (conn *Conn) Query() string {
	return conn.query
}

func (conn *Conn) Outbound() bool {
	return conn.outbound
}
