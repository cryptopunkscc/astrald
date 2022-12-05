package optimizer

import (
	"context"
	"github.com/cryptopunkscc/astrald/infra"
)

type AddrFilterFunc func(addr infra.Addr) error

type FilterDialer struct {
	dialer infra.Dialer
	filter AddrFilterFunc
}

func NewFilterDialer(dialer infra.Dialer, filter AddrFilterFunc) *FilterDialer {
	return &FilterDialer{dialer: dialer, filter: filter}
}

func (t *FilterDialer) Dial(ctx context.Context, addr infra.Addr) (infra.Conn, error) {
	if err := t.filter(addr); err != nil {
		return nil, err
	}

	return t.dialer.Dial(ctx, addr)
}
