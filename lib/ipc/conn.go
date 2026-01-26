package ipc

import "net"

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
