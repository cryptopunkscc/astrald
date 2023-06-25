package node

import (
	"github.com/cryptopunkscc/astrald/net"
	"sync/atomic"
	"time"
)

const (
	StateOpen    = "open"
	StateClosing = "closing"
	StateClosed  = "closed"
)

type Conn struct {
	target        *MonitoredWriter
	caller        *MonitoredWriter
	query         net.Query
	establishedAt time.Time

	remoteClosed atomic.Bool
	localClosed  atomic.Bool
	done         chan struct{}
}

func (conn *Conn) Target() *MonitoredWriter {
	return conn.target
}

func (conn *Conn) Caller() *MonitoredWriter {
	return conn.caller
}

type check interface{ Port() int }

func NewConn(caller *MonitoredWriter, target *MonitoredWriter, query net.Query) *Conn {
	c := &Conn{
		target:        target,
		caller:        caller,
		query:         query,
		done:          make(chan struct{}),
		establishedAt: time.Now(),
	}

	c.target.AfterClose = func(err error) {
		if c.remoteClosed.CompareAndSwap(false, true) {
			c.checkClosed()
		}
	}

	c.caller.AfterClose = func(err error) {
		if c.localClosed.CompareAndSwap(false, true) {
			c.checkClosed()
		}
	}

	return c
}

func (conn *Conn) Query() net.Query {
	return conn.query
}

func (conn *Conn) BytesOut() int {
	if conn.query.Origin() == net.OriginNetwork {
		return conn.caller.Bytes()
	}
	return conn.target.Bytes()
}

func (conn *Conn) BytesIn() int {
	if conn.query.Origin() == net.OriginNetwork {
		return conn.target.Bytes()
	}
	return conn.caller.Bytes()
}

func (conn *Conn) LocalPort() int {
	if p, ok := conn.target.SecureWriteCloser.(check); ok {
		return p.Port()
	}
	return -1
}

func (conn *Conn) RemotePort() int {
	if p, ok := conn.caller.SecureWriteCloser.(check); ok {
		return p.Port()
	}
	return -1
}

func (conn *Conn) Done() <-chan struct{} {
	return conn.done
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

func (conn *Conn) checkClosed() {
	if conn.localClosed.Load() && conn.remoteClosed.Load() {
		close(conn.done)
	}
}
