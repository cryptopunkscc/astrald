package infra

import (
	"context"
	"fmt"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/log"
)

func (i *Infra) Dial(ctx context.Context, addr infra.Addr) (infra.Conn, error) {
	addrStr := fmt.Sprintf("%s:%s", addr.Network(), addr.String())

	i.Logf(log.Debug, "dial %s", addrStr)

	network, found := i.networks[addr.Network()]
	if !found {
		i.Logf(log.Debug, "dial %s error: network unsupported", addrStr)
		return nil, infra.ErrUnsupportedNetwork
	}

	if dialer, ok := network.(infra.Dialer); ok {
		dial, err := dialer.Dial(ctx, addr)

		if err != nil {
			i.Logf(log.Debug, "dial %s error: %v", addrStr, err)
		}

		return dial, err
	}

	i.Logf(log.Debug, "dial %s error: dial unsupported", addrStr)

	return nil, infra.ErrUnsupportedOperation
}
