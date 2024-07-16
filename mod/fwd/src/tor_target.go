package fwd

import (
	"context"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/tor"
	"github.com/cryptopunkscc/astrald/net"
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

func (t *TorTarget) RouteQuery(ctx context.Context, query net.Query, src net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	conn, err := t.tor.Dial(ctx, t.endpoint)
	if err != nil {
		return net.Reject()
	}

	go func() {
		io.Copy(src, conn)
		src.Close()
	}()

	return net.NewSecurePipeWriter(conn, t.identity), nil
}

func (t *TorTarget) String() string {
	return "tor://" + t.endpoint.Address()
}
