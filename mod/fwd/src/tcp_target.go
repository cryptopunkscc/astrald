package fwd

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"io"
	_net "net"
)

type TCPTarget struct {
	identity *astral.Identity
	addr     *_net.TCPAddr
}

func NewTCPTarget(addr string, identiy *astral.Identity) (*TCPTarget, error) {
	var err error
	var tcp = &TCPTarget{identity: identiy}

	tcp.addr, err = _net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}

	return tcp, nil
}

func (t *TCPTarget) RouteQuery(ctx *astral.Context, q *astral.Query, w io.WriteCloser) (io.WriteCloser, error) {
	var dialer = _net.Dialer{}

	conn, err := dialer.DialContext(ctx, "tcp", t.addr.String())
	if err != nil {
		return query.Reject()
	}

	go func() {
		io.Copy(w, conn)
		w.Close()
	}()

	return conn, nil
}

func (t *TCPTarget) String() string {
	return "tcp://" + t.addr.String()
}
