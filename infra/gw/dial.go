package gw

import (
	"context"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/infra"
)

func (g *Gateway) Dial(ctx context.Context, addr Addr) (infra.Conn, error) {
	if len(addr.cookie) == 0 {
		return nil, errors.New("missing cookie")
	}

	rwc, err := g.Query(ctx, addr.gate, PortName)
	if err != nil {
		return nil, fmt.Errorf("gateway query error: %w", err)
	}

	if err := cslq.Encode(rwc, "[c]c", addr.cookie); err != nil {
		return nil, fmt.Errorf("gateway query error: %w", err)
	}

	var res int

	err = cslq.Decode(rwc, "c", &res)
	if err != nil {
		rwc.Close()
		return nil, infra.ErrConnectionRefused
	}

	if res != 1 {
		rwc.Close()
		return nil, infra.ErrConnectionRefused
	}

	return newConn(rwc, addr, true), nil
}
