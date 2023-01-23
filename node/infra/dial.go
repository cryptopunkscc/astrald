package infra

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/infra/bt"
	"github.com/cryptopunkscc/astrald/infra/gw"
	"github.com/cryptopunkscc/astrald/infra/inet"
	"github.com/cryptopunkscc/astrald/infra/tor"
	"strings"
)

func (i *Infra) Dial(ctx context.Context, addr infra.Addr) (conn infra.Conn, err error) {
	// try to repack the address to get the concrete type
	if a, err := i.Unpack(addr.Network(), addr.Pack()); err == nil {
		addr = a
	}

	log.Logv(1, "dial %s %s", log.Em(addr.Network()), addr)

	conn, err = i.dial(ctx, addr)

	if err == nil {
		log.Infov(1, "dial %s %s success", log.Em(addr.Network()), addr)
	} else {
		switch {
		case strings.Contains(err.Error(), "connection refused"),
			strings.Contains(err.Error(), "operation was canceled"),
			strings.Contains(err.Error(), "i/o timeout"),
			errors.Is(err, infra.ErrUnsupportedNetwork),
			errors.Is(err, context.Canceled),
			errors.Is(err, context.DeadlineExceeded):
			log.Errorv(1, "dial %s %s error: %s%s", log.Em(addr.Network()), addr, log.Red(), err.Error())

		default:
			log.Error("dial %s %s error: %s%s", log.Em(addr.Network()), addr, log.Red(), err.Error())
		}
	}

	return
}

func (i *Infra) dial(ctx context.Context, addr infra.Addr) (infra.Conn, error) {
	if _, found := i.networks[addr.Network()]; !found {
		return nil, infra.ErrUnsupportedNetwork
	}

	switch addr := addr.(type) {
	case inet.Addr:
		return i.inet.Dial(ctx, addr)
	case *inet.Addr:
		return i.inet.Dial(ctx, *addr)

	case tor.Addr:
		return i.tor.Dial(ctx, addr)
	case *tor.Addr:
		return i.tor.Dial(ctx, *addr)

	case bt.Addr:
		return i.bluetooth.Dial(ctx, addr)
	case *bt.Addr:
		return i.bluetooth.Dial(ctx, *addr)

	case gw.Addr:
		return i.gateway.Dial(ctx, addr)
	case *gw.Addr:
		return i.gateway.Dial(ctx, *addr)

	default:
		return nil, infra.ErrUnsupportedNetwork
	}
}
