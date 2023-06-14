package link

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/sig"
	"sync/atomic"
)

const (
	StateOpen    = "open"
	StateClosing = "closing"
	StateClosed  = "closed"
)

type Conn struct {
	remoteWriter net.SecureWriteCloser
	localWriter  net.SecureWriteCloser

	query      string
	localPort  int
	remotePort int
	outbound   bool

	activity sig.Activity
	bytesOut int
	bytesIn  int

	remoteClosed atomic.Bool
	localClosed  atomic.Bool
	StateChanged func()
}

func NewConn(localPort int, localWriter net.SecureWriteCloser, remotePort int, remoteWriter net.SecureWriteCloser, query string, outbound bool) *Conn {
	c := &Conn{
		localPort:    localPort,
		localWriter:  localWriter,
		remotePort:   remotePort,
		remoteWriter: remoteWriter,
		query:        query,
		outbound:     outbound,
	}

	if m, ok := c.remoteWriter.(*WriterMonitor); ok {
		m.AfterWrite = func(i int, err error) {
			c.bytesIn += i
		}
		m.AfterClose = func(err error) {
			if c.remoteClosed.CompareAndSwap(false, true) {
				if c.StateChanged != nil {
					c.StateChanged()
				}
			}
		}
	}

	if m, ok := c.localWriter.(*WriterMonitor); ok {
		m.AfterWrite = func(i int, err error) {
			c.bytesOut += i
		}
		m.AfterClose = func(err error) {
			if c.localClosed.CompareAndSwap(false, true) {
				if c.StateChanged != nil {
					c.StateChanged()
				}
			}
		}
	}

	return c
}

func (conn *Conn) LocalIdentity() id.Identity {
	return conn.remoteWriter.RemoteIdentity()
}

func (conn *Conn) RemoteIdentity() id.Identity {
	return conn.localWriter.RemoteIdentity()
}

func (conn *Conn) Query() string {
	return conn.query
}

func (conn *Conn) Outbound() bool {
	return conn.outbound
}

func (conn *Conn) LocalClosed() bool {
	return conn.localClosed.Load()
}

func (conn *Conn) RemoteClosed() bool {
	return conn.remoteClosed.Load()
}

func (conn *Conn) BytesOut() int {
	return conn.bytesOut
}

func (conn *Conn) BytesIn() int {
	return conn.bytesIn
}

func (conn *Conn) LocalPort() int {
	return conn.localPort
}

func (conn *Conn) RemotePort() int {
	return conn.remotePort
}

func (conn *Conn) State() string {
	var c int
	if conn.localClosed.Load() {
		c++
	}
	if conn.remoteClosed.Load() {
		c++
	}
	switch c {
	case 0:
		return StateOpen
	case 1:
		return StateClosing
	case 2:
		return StateClosed
	}
	panic("?")
}
