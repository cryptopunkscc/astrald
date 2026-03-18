package gateway

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/gateway"
)

func (mod *Module) register(ctx *astral.Context, identity *astral.Identity, visibility gateway.Visibility, network string) (gateway.Socket, error) {
	if !mod.canGateway(identity) {
		return gateway.Socket{}, gateway.ErrUnauthorized
	}

	endpoint, err := mod.getGatewayEndpoint(ctx, network)
	if err != nil {
		return gateway.Socket{}, err
	}

	node := &registeredNode{
		Identity:   identity,
		Nonce:      astral.NewNonce(),
		Visibility: visibility,
	}

	old, ok := mod.registeredNodes.Replace(identity.String(), node)
	if ok {
		if err = old.Close(); err != nil {
			mod.log.Error("failed to close old registered node: %v", err)
		}

		targetID := old.Identity.String()
		for _, c := range mod.connectors.Clone() {
			if c.Target.String() == targetID {
				mod.connectors.Remove(c)
				c.Close()
			}
		}
	}

	return gateway.Socket{
		Nonce:    node.Nonce,
		Endpoint: endpoint,
	}, nil
}
