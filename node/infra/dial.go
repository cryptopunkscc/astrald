package infra

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/net"
	"strings"
)

func (i *CoreInfra) Dial(ctx context.Context, e net.Endpoint) (conn net.Conn, err error) {
	// try to repack the address to get the concrete type
	if a, err := i.Unpack(e.Network(), e.Pack()); err == nil {
		e = a
	}

	log.Logv(1, "dial %s %s", e.Network(), e)

	conn, err = i.dial(ctx, e)

	if err == nil {
		log.Infov(1, "dial %s %s success", e.Network(), e)
	} else {
		switch {
		case strings.Contains(err.Error(), "connection refused"),
			strings.Contains(err.Error(), "operation was canceled"),
			strings.Contains(err.Error(), "i/o timeout"),
			errors.Is(err, ErrUnsupportedNetwork),
			errors.Is(err, context.Canceled),
			errors.Is(err, context.DeadlineExceeded):
			log.Errorv(2, "dial %s %s error: %s", e.Network(), e, err)

		default:
			log.Errorv(1, "dial %s %s error: %s", e.Network(), e, err)
		}
	}

	return
}

func (i *CoreInfra) dial(ctx context.Context, addr net.Endpoint) (net.Conn, error) {
	if !i.config.driversContain(addr.Network()) {
		return nil, ErrUnsupportedNetwork
	}

	network, found := i.networkDrivers[addr.Network()]
	if !found {
		return nil, ErrUnsupportedNetwork
	}

	dialer, ok := network.(Dialer)
	if !ok {
		return nil, ErrDialUnsupported
	}

	return dialer.Dial(ctx, addr)
}
