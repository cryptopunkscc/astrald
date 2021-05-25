package inet

import (
	"context"
	"github.com/cryptopunkscc/astrald/node/net"
	goNet "net"
)

type driver struct {
}

func (d *driver) Advertise(ctx context.Context) error {
	return net.ErrUnsupported
}

func (d *driver) Scan(ctx context.Context) (<-chan *net.Ad, error) {
	return nil, net.ErrUnsupported
}

var _ net.Driver = &driver{}

func NewDriver() *driver {
	return &driver{}
}

func (d *driver) Listen(ctx context.Context) (<-chan net.Conn, error) {
	return nil, net.ErrUnsupported
}

func (d *driver) Network() string {
	return "inet"
}

func (d *driver) Dial(_ context.Context, ep net.Addr) (net.Conn, error) {
	tcpConn, err := goNet.Dial("tcp", ep.String())
	if err != nil {
		return nil, err
	}

	return net.WrapConn(tcpConn, true), nil
}
