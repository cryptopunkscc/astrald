package gw

import (
	"context"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/enc"
	"github.com/cryptopunkscc/astrald/infra"
)

func (g Gateway) Dial(ctx context.Context, addr infra.Addr) (infra.Conn, error) {
	a, ok := addr.(Addr)
	if !ok {
		return nil, infra.ErrUnsupportedAddress
	}

	if len(a.cookie) == 0 {
		return nil, errors.New("missing cookie")
	}

	rwc, err := g.Query(ctx, a.gate, PortName)
	if err != nil {
		return nil, fmt.Errorf("gateway query error: %w", err)
	}

	if err := enc.WriteL8String(rwc, a.cookie); err != nil {
		return nil, fmt.Errorf("gateway query error: %w", err)
	}

	res, err := enc.ReadUint8(rwc)
	if err != nil {
		rwc.Close()
		return nil, infra.ErrConnectionRefused
	}

	if res != 1 {
		rwc.Close()
		return nil, infra.ErrConnectionRefused
	}

	return newConn(rwc, a, true), nil
}
