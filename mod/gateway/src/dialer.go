package gateway

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/exonet"
)

type Dialer struct {
	node astral.Node
}

func NewDialer(node astral.Node) *Dialer {
	return &Dialer{node: node}
}

func (dialer *Dialer) Dial(ctx context.Context, endpoint exonet.Endpoint) (exonet.Conn, error) {
	e, err := Unpack(endpoint.Pack())
	if err != nil {
		return nil, err
	}

	if e.GatewayID.IsEqual(dialer.node.Identity()) {
		return nil, ErrSelfGateway
	}

	var q = astral.NewQuery(dialer.node.Identity(), e.GatewayID, RouteServiceName+"."+e.TargetID.String())

	conn, err := query.Route(ctx, dialer.node, q)
	if err != nil {
		return nil, err
	}

	return newConn(
		conn,
		NewEndpoint(dialer.node.Identity(), dialer.node.Identity()),
		e,
		true,
	), err
}
