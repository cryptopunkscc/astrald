package infra

import (
	"context"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/log"
)

func (i *Infra) Dial(ctx context.Context, addr infra.Addr) (infra.Conn, error) {
	i.Logf(log.Debug, "dial %s", addr.String())

	network, found := i.networks[addr.Network()]
	if !found {
		i.Logf(log.Debug, "dial error: network unsupported")
		return nil, infra.ErrUnsupportedNetwork
	}

	if dialer, ok := network.(infra.Dialer); ok {
		dial, err := dialer.Dial(ctx, addr)

		if err != nil {
			i.Logf(log.Debug, "dial error: %v", err)
		}

		return dial, err
	}

	i.Logf(log.Debug, "dial error: network %s does not support dialing", network.Name())

	return nil, infra.ErrUnsupportedOperation
}
