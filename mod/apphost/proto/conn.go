package proto

import (
	"github.com/cryptopunkscc/astrald/cslq"
	"net"
)

type Conn struct {
	net.Conn
}

func NewConn(conn net.Conn) *Conn {
	return &Conn{Conn: conn}
}

func (c *Conn) ReadMsg(msg interface{}) error {
	return cslq.Decode(c, "v", msg)
}

func (c *Conn) WriteMsg(msg interface{}) error {
	return cslq.Encode(c, "v", msg)
}

func (c *Conn) ReadErr() error {
	var code int
	if err := cslq.Decode(c, ErrorCSLQ, &code); err != nil {
		return err
	}
	return ErrorCode(code)
}

func (c *Conn) WriteErr(e APIError) error {
	if e == nil {
		return cslq.Encode(c, ErrorCSLQ, Success)
	}
	return cslq.Encode(c, ErrorCSLQ, e.Code())
}
