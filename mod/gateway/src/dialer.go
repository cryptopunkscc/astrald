package gateway

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
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

	if e.Gate().IsEqual(dialer.node.Identity()) {
		return nil, ErrSelfGateway
	}

	var query = astral.NewQuery(dialer.node.Identity(), e.Gate(), RouteServiceName+"."+e.Target().PublicKeyHex())

	conn, err := astral.Route(ctx, dialer.node, query)
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
