package fwd

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/infra/drivers/tor"
	"io"
)

type TorTarget struct {
	tor      *tor.Driver
	identity id.Identity
	endpoint tor.Endpoint
}

func NewTorTarget(drv *tor.Driver, addr string, identiy id.Identity) (*TorTarget, error) {
	var err error
	var t = &TorTarget{
		identity: identiy,
		tor:      drv,
	}

	t.endpoint, err = tor.Parse(addr)
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
	return "tor://" + t.endpoint.String()
}
