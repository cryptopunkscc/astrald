package net

import "net"

// WrappedConn represents a net.Conn over a TCP/IPv4 connection
type WrappedConn struct {
	net.Conn
	outbound bool
}

// WrapConn wraps a standard net.Conn into a astral's net.Conn with the addition of boundness
func WrapConn(conn net.Conn, outbound bool) *WrappedConn {
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
