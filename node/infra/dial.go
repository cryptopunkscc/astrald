package infra

import (
	"context"
	"github.com/cryptopunkscc/astrald/infra"
)

func (i *Infra) Dial(ctx context.Context, addr infra.Addr) (infra.Conn, error) {
	network, found := i.networks[addr.Network()]
	if !found {
		return nil, infra.ErrUnsupportedNetwork
	}

	if dialer, ok := network.(infra.Dialer); ok {
		return dialer.Dial(ctx, addr)
	}

	return nil, infra.ErrUnsupportedOperation
}
