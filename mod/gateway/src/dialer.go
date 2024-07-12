package gateway

import (
	"context"
	"github.com/cryptopunkscc/astrald/net"
	node2 "github.com/cryptopunkscc/astrald/node"
)

type Dialer struct {
	node node2.Node
}

func NewDialer(node node2.Node) *Dialer {
	return &Dialer{node: node}
}

func (dialer *Dialer) Dial(ctx context.Context, endpoint net.Endpoint) (net.Conn, error) {
	e, err := Unpack(endpoint.Pack())
	if err != nil {
		return nil, err
	}

	if e.Gate().IsEqual(dialer.node.Identity()) {
		return nil, ErrSelfGateway
	}

	var query = net.NewQuery(dialer.node.Identity(), e.Gate(), RouteServiceName+"."+e.Target().PublicKeyHex())

	conn, err := net.Route(ctx, dialer.node.Router(), query)
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
