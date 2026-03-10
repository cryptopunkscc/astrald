package gateway

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/astrald"
	"github.com/cryptopunkscc/astrald/mod/gateway"
	gatewayClient "github.com/cryptopunkscc/astrald/mod/gateway/client"
)

func (mod *Module) bindToGateway(ctx *astral.Context, gatewayID *astral.Identity, visibility gateway.Visibility) {
	client := gatewayClient.New(gatewayID, astrald.Default())

	socket, err := client.Bind(ctx, visibility)
	if err != nil {
		mod.log.Error("bind to %v: %v", gatewayID, err)
		return
	}

	newReceiverConnPool(mod, socket).Run(ctx)
}
