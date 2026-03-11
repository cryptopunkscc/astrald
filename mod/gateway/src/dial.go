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

	client := gatewayClient.New(gwEndpoint.GatewayID, astrald.Default())

	socket, err := client.Connect(ctx.IncludeZone(astral.ZoneNetwork), gwEndpoint.TargetID)
	if err != nil {
		return nil, err
	}

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
