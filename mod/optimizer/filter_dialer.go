package optimizer

import (
	"context"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/infra"
)

type AddrFilterFunc func(net.Endpoint) error

type FilterDialer struct {
	dialer infra.Dialer
	filter AddrFilterFunc
}

func NewFilterDialer(dialer infra.Dialer, filter AddrFilterFunc) *FilterDialer {
	return &FilterDialer{dialer: dialer, filter: filter}
}

func (t *FilterDialer) Dial(ctx context.Context, e net.Endpoint) (net.Conn, error) {
	if err := t.filter(e); err != nil {
		return nil, err
	}

	return t.dialer.Dial(ctx, e)
}
