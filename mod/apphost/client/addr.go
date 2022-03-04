package astral

import "net"

type Conn struct {
	net.Conn
	remoteAddr Addr
}

type Addr struct {
	address string
}

func (a Addr) String() string {
	return a.address
}

func (a Addr) Network() string {
	return "astral"
}

func (conn Conn) RemoteAddr() net.Addr {
	return conn.remoteAddr
}
