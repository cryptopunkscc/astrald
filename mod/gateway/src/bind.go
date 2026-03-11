package gateway

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/gateway"
)

func (mod *Module) bind(ctx *astral.Context, identity *astral.Identity, visibility gateway.Visibility, network string) (*gateway.Socket, error) {
	if !mod.canGateway(identity) {
		return nil, gateway.ErrUnauthorized
	}

	endpoint, err := mod.getGatewayEndpoint(ctx, network)
	if err != nil {
		return nil, err
	}

	nonce := astral.NewNonce()

	client := &client{
		Identity:   identity,
		Nonce:      nonce,
		Visibility: visibility,
		Target:     nil, // its a binder
	}

	oldClient, ok := mod.binders.Replace(identity.String(), client)
	if ok {
		err = oldClient.Close()
		if err != nil {
			mod.log.Error("failed to close oldClient client: %v", err)
		}
	}

	return &gateway.Socket{
		Nonce:    nonce,
		Endpoint: endpoint,
	}, nil
}
