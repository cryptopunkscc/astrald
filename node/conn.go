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

var nextConnID atomic.Int64

type Conn struct {
	id            int64
	target        *MonitoredWriter
	caller        *MonitoredWriter
	query         net.Query
	hints         net.Hints
	establishedAt time.Time

	targetClosed atomic.Bool
	callerClosed atomic.Bool
	done         chan struct{}
}

func NewConn(caller *MonitoredWriter, target *MonitoredWriter, query net.Query, hints net.Hints) *Conn {
	c := &Conn{
		id:            nextConnID.Add(1),
		target:        target,
		caller:        caller,
		query:         query,
		hints:         hints,
		done:          make(chan struct{}),
		establishedAt: time.Now(),
	}

	c.target.AfterClose = func(err error) {
		if c.targetClosed.CompareAndSwap(false, true) {
			c.checkClosed()
		}
	}

	c.caller.AfterClose = func(err error) {
		if c.callerClosed.CompareAndSwap(false, true) {
			c.checkClosed()
		}
	}

	return c
}

func (conn *Conn) ID() int {
	return int(conn.id)
}

func (conn *Conn) Target() *MonitoredWriter {
	return conn.target
}

func (conn *Conn) Caller() *MonitoredWriter {
	return conn.caller
}

func (conn *Conn) Query() net.Query {
	return conn.query
}

func (conn *Conn) Hints() net.Hints {
	return conn.hints
}

func (conn *Conn) BytesOut() int {
	return conn.target.Bytes()
}

func (conn *Conn) BytesIn() int {
	return conn.caller.Bytes()
}

func (conn *Conn) Done() <-chan struct{} {
	return conn.done
}

func (conn *Conn) State() string {
	var c int
	if conn.callerClosed.Load() {
		c++
	}
	if conn.targetClosed.Load() {
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
	if conn.callerClosed.Load() && conn.targetClosed.Load() {
		close(conn.done)
	}
}
