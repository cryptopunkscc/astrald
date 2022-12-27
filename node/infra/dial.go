package infra

import (
	"context"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/infra/bt"
	"github.com/cryptopunkscc/astrald/infra/gw"
	"github.com/cryptopunkscc/astrald/infra/inet"
	"github.com/cryptopunkscc/astrald/infra/tor"
	"github.com/cryptopunkscc/astrald/log"
)

func (i *Infra) Dial(ctx context.Context, addr infra.Addr) (conn infra.Conn, err error) {
	i.Logf(log.Debug, "dial %s %s", addr.Network(), addr.String())

	conn, err = i.dial(ctx, addr)

	if err == nil {
		i.Logf(log.Debug, "dial %s %s success", addr.Network(), addr.String())
	} else {
		i.Logf(log.Debug, "dial %s %s error: %s", addr.Network(), addr.String(), err.Error())
	}

	return
}

func (i *Infra) dial(ctx context.Context, addr infra.Addr) (infra.Conn, error) {
	var err error

	// if it's a generic addr, try to decode
	if _, ok := addr.(*infra.GenericAddr); ok {
		addr, err = i.Unpack(addr.Network(), addr.Pack())
		if err != nil {
			return nil, err
		}
	}

	if _, found := i.networks[addr.Network()]; !found {
		return nil, infra.ErrUnsupportedNetwork
	}

	switch addr := addr.(type) {
	case inet.Addr:
		return i.inet.Dial(ctx, addr)
	case *inet.Addr:
		return i.inet.Dial(ctx, *addr)

	case tor.Addr:
		if i.tor == nil {
			panic("tor is nil")
		}
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
