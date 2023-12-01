package infra

import (
	"context"
	"github.com/cryptopunkscc/astrald/net"
)

func (infra *CoreInfra) Dial(ctx context.Context, addr net.Endpoint) (net.Conn, error) {
	infra.mu.RLock()
	defer infra.mu.RUnlock()

	if dialer, found := infra.dialers[addr.Network()]; found {
		return dialer.Dial(ctx, addr)
	}

	return nil, ErrUnsupportedNetwork
}

func (infra *CoreInfra) SetDialer(network string, dialer Dialer) {
	infra.mu.Lock()
	defer infra.mu.Unlock()

	if dialer == nil {
		delete(infra.dialers, network)
	} else {
		infra.dialers[network] = dialer
	}
}
