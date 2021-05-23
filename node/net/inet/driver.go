package inet

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/node/net"
	_net "net"
)

type driver struct {
}

func (d *driver) Advertise(ctx context.Context) error {
	return errors.New("advertising not supported")
}

func (d *driver) Scan(ctx context.Context) (<-chan *net.Ad, error) {
	return nil, errors.New("scan not supported")
}

var _ net.Driver = &driver{}

func NewDriver() *driver {
	return &driver{}
}

func (d *driver) Listen(ctx context.Context) (<-chan net.Conn, error) {
	return nil, errors.New("unsupported")
}

func (d *driver) Network() string {
	return "inet"
}

func (d *driver) Dial(_ context.Context, ep net.Addr) (net.Conn, error) {
	tcpConn, err := _net.Dial("tcp", ep.String())
	if err != nil {
		return nil, err
	}

	return net.WrapConn(tcpConn, true), nil
}
