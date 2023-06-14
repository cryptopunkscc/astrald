package link

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mux"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/sig"
	"io"
	"sync"
	"sync/atomic"
	"time"
)

var _ net.SecureConn = &Conn{}

type Conn struct {
	remoteWriter net.SecureWriteCloser
	localWriter  net.SecureWriteCloser

	localPort      int
	reader         io.Reader
	writer         io.WriteCloser
	fallbackReader io.Reader
	query          string
	link           *Link
	outbound       bool
	activity       sig.Activity
	done           chan struct{}
	closed         atomic.Bool
	err            error
	wmu            sync.Mutex // writer mutex
	frmu           sync.Mutex // fallback reader mutex
	lmu            sync.Mutex // link mutex
}

func (conn *Conn) Link() *Link {
	return conn.link
}

func (conn *Conn) LocalEndpoint() net.Endpoint {
	return nil
}

func (conn *Conn) RemoteEndpoint() net.Endpoint {
	return nil
}

func (conn *Conn) RemoteIdentity() id.Identity {
	return conn.link.RemoteIdentity()
}

func (conn *Conn) LocalIdentity() id.Identity {
	return conn.link.LocalIdentity()
}

func (conn *Conn) Read(p []byte) (n int, err error) {
	if conn.closed.Load() {
		return 0, conn.err
	}

	defer conn.activity.Touch()

	n, err = conn.reader.Read(p)
	if err == nil {
		conn.link.Health().Check()
		return
	}

	// try to switch to fallback reader and try reading again
	if conn.switchToFallbackReader() {
		return conn.Read(p)
	}

	conn.closeWithError(err)

	return n, conn.err
}

func (conn *Conn) Write(p []byte) (n int, err error) {
	conn.wmu.Lock()
	defer conn.wmu.Unlock()

	if conn.closed.Load() {
		return 0, conn.err
	}

	defer conn.activity.Touch()

	conn.link.Health().Check()

	n, err = conn.writer.Write(p)
	if err != nil {
		conn.closeWithError(err)
	}

	return n, conn.err
}

func (conn *Conn) Close() error {
	return conn.closeWithError(ErrConnClosed)
}

func (conn *Conn) SetFallbackReader(r io.Reader) {
	conn.frmu.Lock()
	defer conn.frmu.Unlock()

	conn.fallbackReader = r
}

func (conn *Conn) SetWriter(w io.WriteCloser) error {
	conn.wmu.Lock()
	defer conn.wmu.Unlock()

	if conn.closed.Load() {
		return conn.err
	}

	conn.writer.Close()
	conn.writer = w
	return nil
}

func (conn *Conn) Attach(link *Link) {
	conn.lmu.Lock()
	defer conn.lmu.Unlock()

	if conn.link == link {
		return
	}

	if conn.link != nil {
		conn.link.remove(conn)
	}

	conn.link = link

	if conn.link != nil {
		conn.link.add(conn)
	}
}

func (conn *Conn) LocalPort() int {
	return conn.localPort
}

// RemotePort returns the remote port on the multiplexer, or -1 if the connection is not multiplexed.
func (conn *Conn) RemotePort() int {
	if w, ok := conn.writer.(*mux.FrameWriter); ok {
		return w.Port()
	}
	return -1
}

func (conn *Conn) Outbound() bool {
	return conn.outbound
}

func (conn *Conn) Source() any {
	return conn.link
}

func (conn *Conn) Query() string {
	return conn.query
}

func (conn *Conn) Idle() time.Duration {
	return conn.activity.Idle()
}

func (conn *Conn) Done() <-chan struct{} {
	return conn.done
}

func (conn *Conn) closeWithError(e error) error {
	if conn.closed.CompareAndSwap(false, true) {
		conn.writer.Close()
		conn.err = e
		close(conn.done)
		conn.link.Events().Emit(EventConnClosed{Conn: conn})
		conn.link.remove(conn)
	}
	return nil
}

// switchToFallback switches the reader to the fallback if one is set.
// Returns true if switch was successful, false otherwise.
func (conn *Conn) switchToFallbackReader() bool {
	conn.frmu.Lock()
	defer conn.frmu.Unlock()

	if conn.fallbackReader == nil {
		return false
	}

	conn.reader = conn.fallbackReader
	conn.fallbackReader = nil
	return true
}
