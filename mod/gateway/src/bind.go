package gateway

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/gateway"
)

func (mod *Module) bind(ctx *astral.Context, identity *astral.Identity, visibility gateway.Visibility, network string) (gateway.Socket, error) {
	if !mod.canGateway(identity) {
		return gateway.Socket{}, gateway.ErrUnauthorized
	}

	endpoint, err := mod.getGatewayEndpoint(ctx, network)
	if err != nil {
		return gateway.Socket{}, err
	}

	newBinder := &binder{
		Identity:   identity,
		Nonce:      astral.NewNonce(),
		Visibility: visibility,
	}

	oldBinder, ok := mod.binders.Replace(identity.String(), newBinder)
	if ok {
		if err = oldBinder.Close(); err != nil {
			mod.log.Error("failed to close old binder: %v", err)
		}

		targetID := oldBinder.Identity.String()
		for _, c := range mod.connectors.Clone() {
			if c.Target.String() == targetID {
				mod.connectors.Remove(c)
				c.Close()
			}
		}
	}

	return gateway.Socket{
		Nonce:    newBinder.Nonce,
		Endpoint: endpoint,
	}, nil
}
