package gateway

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/astrald"
	"github.com/cryptopunkscc/astrald/lib/query"
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

	// Fast path: socket-based (raw TCP piped through gateway)
	if socket, err := client.Connect(ctx.IncludeZone(astral.ZoneNetwork), gwEndpoint.TargetID); err == nil {
		if conn, err := mod.Exonet.Dial(ctx, socket.Endpoint); err == nil {
			if _, err = socket.Nonce.WriteTo(conn); err == nil {
				return &gwConn{
					ReadWriteCloser: conn,
					local:           conn.LocalEndpoint(),
					remote:          gwEndpoint,
					outbound:        conn.Outbound(),
				}, nil
			}
			conn.Close()
		}
	}

	// Slow path: link-based (route query through existing astral links)
	mod.log.Logv(1, "socket path unavailable, trying link path to %v via %v", gwEndpoint.TargetID, gwEndpoint.GatewayID)

	q := &astral.Query{
		Nonce:  astral.NewNonce(),
		Caller: mod.node.Identity(),
		Target: gwEndpoint.GatewayID,
		Query:  gateway.MethodRoute + "." + gwEndpoint.TargetID.String(),
	}

	conn, err := query.Route(ctx, mod.node, q)
	if err != nil {
		return nil, err
	}

	return &gwConn{
		ReadWriteCloser: conn,
		local:           gateway.NewEndpoint(mod.node.Identity(), mod.node.Identity()),
		remote:          gwEndpoint,
		outbound:        true,
	}, nil
}
