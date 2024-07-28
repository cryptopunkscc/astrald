package fwd

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/tor"
	"io"
)

type TorTarget struct {
	tor      tor.Module
	identity id.Identity
	endpoint exonet.Endpoint
}

func NewTorTarget(drv tor.Module, addr string, identiy id.Identity) (*TorTarget, error) {
	var err error
	var t = &TorTarget{
		identity: identiy,
		tor:      drv,
	}

	t.endpoint, err = drv.Parse("tor", addr)
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (t *TorTarget) RouteQuery(ctx context.Context, query *astral.Query, caller io.WriteCloser, hints astral.Hints) (io.WriteCloser, error) {
	conn, err := t.tor.Dial(ctx, t.endpoint)
	if err != nil {
		return astral.Reject()
	}

	go func() {
		io.Copy(caller, conn)
		caller.Close()
	}()

	return conn, nil
}

func (t *TorTarget) String() string {
	return "tor://" + t.endpoint.Address()
}
