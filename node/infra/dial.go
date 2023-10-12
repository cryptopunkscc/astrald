package infra

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/net"
	"strings"
)

func (infra *CoreInfra) Dial(ctx context.Context, e net.Endpoint) (conn net.Conn, err error) {
	// try to repack the address to get the concrete type
	if a, err := infra.Unpack(e.Network(), e.Pack()); err == nil {
		e = a
	}

	infra.log.Logv(2, "dial %s %s", e.Network(), e)

	conn, err = infra.dial(ctx, e)

	if err == nil {
		infra.log.Infov(2, "dial %s %s success", e.Network(), e)
	} else {
		switch {
		case strings.Contains(err.Error(), "connection refused"),
			strings.Contains(err.Error(), "operation was canceled"),
			strings.Contains(err.Error(), "i/o timeout"),
			errors.Is(err, ErrUnsupportedNetwork),
			errors.Is(err, context.Canceled),
			errors.Is(err, context.DeadlineExceeded):
			infra.log.Errorv(2, "dial %s %s error: %s", e.Network(), e, err)

		default:
			infra.log.Errorv(1, "dial %s %s error: %s", e.Network(), e, err)
		}
	}

	return
}

func (infra *CoreInfra) AddDialer(network string, d Dialer) error {
	infra.mu.Lock()
	defer infra.mu.Unlock()

	if _, found := infra.dialers[network]; found {
		return errors.New("already added")
	}

	infra.dialers[network] = d

	return nil
}

func (infra *CoreInfra) RemoveDialer(network string) error {
	infra.mu.Lock()
	defer infra.mu.Unlock()

	if _, found := infra.dialers[network]; !found {
		return errors.New("not found")
	}

	delete(infra.dialers, network)

	return nil
}

func (infra *CoreInfra) dial(ctx context.Context, addr net.Endpoint) (net.Conn, error) {
	if dialer, found := infra.dialers[addr.Network()]; found {
		return dialer.Dial(ctx, addr)
	}

	if !infra.config.driversContain(addr.Network()) {
		return nil, ErrUnsupportedNetwork
	}

	network, found := infra.networkDrivers[addr.Network()]
	if !found {
		return nil, ErrUnsupportedNetwork
	}

	dialer, ok := network.(Dialer)
	if !ok {
		return nil, ErrDialUnsupported
	}

	return dialer.Dial(ctx, addr)
}
