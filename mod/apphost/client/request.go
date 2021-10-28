package astral

import (
	"github.com/cryptopunkscc/astrald/mod/apphost/proto"
	"io"
	"net"
)

type Request struct {
	caller string
	query  string
	raw    net.Conn
	conn   *proto.Conn
}

func (request Request) Caller() string {
	return request.caller
}

func (request Request) Query() string {
	return request.query
}

func (request *Request) Accept() (io.ReadWriteCloser, error) {
	return request.raw, request.conn.OK()
}

func (request *Request) Reject() {
	request.conn.Error("rejected")
}
