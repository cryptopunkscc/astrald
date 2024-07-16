package fwd

import (
	"context"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/net"
	"io"
	_net "net"
)

type TCPTarget struct {
	identity id.Identity
	addr     *_net.TCPAddr
}

func NewTCPTarget(addr string, identiy id.Identity) (*TCPTarget, error) {
	var err error
	var tcp = &TCPTarget{identity: identiy}

	tcp.addr, err = _net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}

	return tcp, nil
}

func (t *TCPTarget) RouteQuery(ctx context.Context, query net.Query, src net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	var dialer = _net.Dialer{}

	conn, err := dialer.DialContext(ctx, "tcp", t.addr.String())
	if err != nil {
		return net.Reject()
	}

	go func() {
		io.Copy(src, conn)
		src.Close()
	}()

	return net.NewSecurePipeWriter(conn, t.identity), nil
}

func (t *TCPTarget) String() string {
	return "tcp://" + t.addr.String()
}
