package gateway

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/astrald"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/gateway"
	gatewayClient "github.com/cryptopunkscc/astrald/mod/gateway/client"
)

var _ exonet.Dialer = &Module{}

func (mod *Module) Dial(ctx *astral.Context, endpoint exonet.Endpoint) (exonet.Conn, error) {
	if endpoint.Network() != NetworkName {
		return nil, exonet.ErrUnsupportedNetwork
	}

	gwEndpoint, ok := endpoint.(*gateway.Endpoint)
	if !ok {
		return nil, exonet.ErrUnsupportedNetwork
	}

	if gwEndpoint.GatewayID.IsEqual(mod.node.Identity()) {
		return nil, gateway.ErrInvalidGateway
	}

	client := gatewayClient.New(gwEndpoint.GatewayID, astrald.Default())

	socket, err := client.Connect(ctx.IncludeZone(astral.ZoneNetwork), gwEndpoint.TargetID)
	if err != nil {
		return nil, err
	}

	// todo: if we cannot obtain socket we should try to connect to gateway and route to target over link

	conn, err := mod.Exonet.Dial(ctx, socket.Endpoint)
	if err != nil {
		return nil, err
	}

	if _, err := socket.Nonce.WriteTo(conn); err != nil {
		conn.Close()
		return nil, err
	}

	return &gwConn{
		Conn:   conn,
		remote: gwEndpoint,
	}, nil
}
