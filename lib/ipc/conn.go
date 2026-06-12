package ipc

import "net"

// Conn wraps a net.Conn with the IPC protocol and address used to establish it.
type Conn struct {
	net.Conn
	protocol string
	addr     string
}

func (conn *Conn) Protocol() string {
	return conn.protocol
}

func (conn *Conn) Endpoint() string {
	return conn.protocol + ":" + conn.addr
}
