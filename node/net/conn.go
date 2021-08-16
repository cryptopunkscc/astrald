package net

import (
	"io"
	go_net "net"
)

// Conn represents a generic interface for any type of bidirectional stream of bytes that preserves order
// and gaurantees data integrity. No authentication or encryption is guaranteed. Identity of the other party is
// unknown.
type Conn interface {
	io.ReadWriteCloser // Basic IO operations
	Outbound() bool    // Returns true if we are the active party, false otherwise
	RemoteAddr() Addr  // Returns the other party's network address if knwon, nil otherwise
}

// WrappedConn wraps a standard go Conn to satisfy astral Conn interface
type WrappedConn struct {
	go_net.Conn
	outbound bool
}

var _ Conn = &WrappedConn{}

// WrapConn wraps a standard net.Conn into a astral's net.Conn with the addition of boundness
func WrapConn(conn go_net.Conn, outbound bool) *WrappedConn {
	return &WrappedConn{
		Conn:     conn,
		outbound: outbound,
	}
}

var _ Conn = &WrappedConn{}

func (conn *WrappedConn) RemoteAddr() Addr {
	return conn.Conn.RemoteAddr()
}

func (conn *WrappedConn) Outbound() bool {
	return conn.outbound
}
