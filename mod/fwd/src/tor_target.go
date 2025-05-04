package fwd

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/tor"
	"io"
)

type TorTarget struct {
	tor      tor.Module
	identity *astral.Identity
	endpoint exonet.Endpoint
}

func NewTorTarget(drv tor.Module, addr string, identiy *astral.Identity) (*TorTarget, error) {
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

func (t *TorTarget) RouteQuery(ctx *astral.Context, q *astral.Query, w io.WriteCloser) (io.WriteCloser, error) {
	conn, err := t.tor.Dial(ctx, t.endpoint)
	if err != nil {
		return query.Reject()
	}

	go func() {
		io.Copy(w, conn)
		w.Close()
	}()

	return conn, nil
}

func (t *TorTarget) String() string {
	return "tor://" + t.endpoint.Address()
}
