package gateway

import (
	"context"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node"
)

type Dialer struct {
	node node.Node
}

func NewDialer(node node.Node) *Dialer {
	return &Dialer{node: node}
}

func (dialer *Dialer) Dial(ctx context.Context, endpoint net.Endpoint) (net.Conn, error) {
	e, err := Unpack(endpoint.Pack())
	if err != nil {
		return nil, err
	}

	var query = net.NewQuery(dialer.node.Identity(), e.Gate(), RouteServiceName+"?"+e.Target().PublicKeyHex())

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
