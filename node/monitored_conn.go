package node

import (
	"github.com/cryptopunkscc/astrald/net"
	"sync/atomic"
	"time"
)

const (
	StateRouting = "routing"
	StateOpen    = "open"
	StateClosing = "closing"
	StateClosed  = "closed"
)

var nextConnID atomic.Int64

type MonitoredConn struct {
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

func NewMonitoredConn(caller *MonitoredWriter, target *MonitoredWriter, query net.Query, hints net.Hints) *MonitoredConn {
	conn := &MonitoredConn{
		id:            nextConnID.Add(1),
		query:         query,
		hints:         hints,
		done:          make(chan struct{}),
		establishedAt: time.Now(),
	}

	conn.SetCaller(caller)
	conn.SetTarget(target)

	return conn
}

func (conn *MonitoredConn) SetTarget(target *MonitoredWriter) {
	conn.target = target
	if target != nil {
		target.AfterClose = func(err error) {
			if conn.targetClosed.CompareAndSwap(false, true) {
				conn.checkClosed()
			}
		}
	}
}

func (conn *MonitoredConn) SetCaller(caller *MonitoredWriter) {
	conn.caller = caller
	if caller != nil {
		caller.AfterClose = func(err error) {
			if conn.callerClosed.CompareAndSwap(false, true) {
				conn.checkClosed()
			}
		}
	}
}

func (conn *MonitoredConn) ID() int {
	return int(conn.id)
}

func (conn *MonitoredConn) Target() *MonitoredWriter {
	return conn.target
}

func (conn *MonitoredConn) Caller() *MonitoredWriter {
	return conn.caller
}

func (conn *MonitoredConn) Query() net.Query {
	return conn.query
}

func (conn *MonitoredConn) SetQuery(query net.Query) {
	conn.query = query
}

func (conn *MonitoredConn) Hints() net.Hints {
	return conn.hints
}

func (conn *MonitoredConn) BytesOut() int {
	if conn.target == nil {
		return 0
	}
	return conn.target.Bytes()
}

func (conn *MonitoredConn) BytesIn() int {
	if conn.caller == nil {
		return 0
	}
	return conn.caller.Bytes()
}

func (conn *MonitoredConn) Done() <-chan struct{} {
	return conn.done
}

func (conn *MonitoredConn) State() string {
	if conn.target == nil {
		return StateRouting
	}
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

func (conn *MonitoredConn) checkClosed() {
	if conn.callerClosed.Load() && conn.targetClosed.Load() {
		close(conn.done)
	}
}
