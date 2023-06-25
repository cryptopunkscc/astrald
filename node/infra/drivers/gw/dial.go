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

func (drv *Driver) Dial(ctx context.Context, e net.Endpoint) (net.Conn, error) {
	e, err := drv.Unpack(e.Network(), e.Pack())
	if err != nil {
		return nil, err
	}

	endpoint := e.(Endpoint)

	if endpoint.gate.IsEqual(drv.infra.Node().Identity()) {
		return nil, errors.New("cannot use self as a gateway")
	}

	rwc, err := net.Route(ctx, drv.infra.Node().Router(), net.NewQuery(drv.infra.Node().Identity(), endpoint.Gate(), ServiceName))
	if err != nil {
		return nil, fmt.Errorf("gateway query error: %w", err)
	}

	if err := cslq.Encode(rwc, "[c]c", endpoint.target.PublicKeyHex()); err != nil {
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

	return newConn(rwc, endpoint, true), nil
}
