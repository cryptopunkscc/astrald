package gateway

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/gateway"
)

type Dialer struct {
	node astral.Node
}

func NewDialer(node astral.Node) *Dialer {
	return &Dialer{node: node}
}

func (dialer *Dialer) Dial(ctx *astral.Context, endpoint exonet.Endpoint) (exonet.Conn, error) {
	e, err := Unpack(endpoint.Pack())
	if err != nil {
		return nil, err
	}

	if e.GatewayID.IsEqual(dialer.node.Identity()) {
		return nil, ErrInvalidGateway
	}

	var q = astral.NewQuery(dialer.node.Identity(), e.GatewayID, RouteServiceName+"."+e.TargetID.String())

	conn, err := query.Route(ctx, dialer.node, q)
	if err != nil {
		return nil, err
	}

	return newConn(
		conn,
		gateway.NewEndpoint(dialer.node.Identity(), dialer.node.Identity()),
		e,
		true,
	), err
}
