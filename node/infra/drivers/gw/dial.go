package gw

import (
	"context"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/infra"
)

var _ infra.Dialer = &Driver{}

func (drv *Driver) Dial(ctx context.Context, addr net.Endpoint) (net.Conn, error) {
	addr, err := drv.Unpack(addr.Network(), addr.Pack())
	if err != nil {
		return nil, err
	}

	gwAddr := addr.(Endpoint)

	if gwAddr.gate.IsEqual(drv.infra.Node().Identity()) {
		return nil, errors.New("cannot use self as a gateway")
	}

	rwc, err := drv.infra.Node().Query(ctx, gwAddr.gate, PortName)
	if err != nil {
		return nil, fmt.Errorf("gateway query error: %w", err)
	}

	if err := cslq.Encode(rwc, "[c]c", gwAddr.target.PublicKeyHex()); err != nil {
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

	return newConn(rwc, gwAddr, true), nil
}
