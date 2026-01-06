package astrald

import (
	"sync/atomic"

	"github.com/cryptopunkscc/astrald/astral"
)

type ConnMonitor struct {
	OnClose      func()
	OnReadError  func(error)
	OnWriteError func(error)
	conn         astral.Conn
	query        *astral.Query
	bytesRead    atomic.Uint64
	bytesWritten atomic.Uint64
}

var _ astral.Conn = &ConnMonitor{}

func (monitor *ConnMonitor) Read(b []byte) (n int, err error) {
	n, err = monitor.conn.Read(b)
	monitor.bytesRead.Add(uint64(n))
	if err != nil {
		monitor.OnReadError(err)
	}
	return
}

func (monitor *ConnMonitor) Write(b []byte) (n int, err error) {
	n, err = monitor.conn.Write(b)
	monitor.bytesWritten.Add(uint64(n))
	if err != nil {
		monitor.OnWriteError(err)
	}
	return
}

func (monitor *ConnMonitor) Close() error {
	if monitor.OnClose != nil {
		defer monitor.OnClose()
	}

	return monitor.conn.Close()
}

func (monitor *ConnMonitor) LocalIdentity() *astral.Identity {
	return monitor.LocalIdentity()
}

func (monitor *ConnMonitor) RemoteIdentity() *astral.Identity {
	return monitor.RemoteIdentity()
}

func (monitor *ConnMonitor) BytesRead() astral.Size {
	return astral.Size(monitor.bytesRead.Load())
}

func (monitor *ConnMonitor) BytesWritten() astral.Size {
	return astral.Size(monitor.bytesWritten.Load())
}

func (monitor *ConnMonitor) Query() *astral.Query {
	return monitor.query
}
