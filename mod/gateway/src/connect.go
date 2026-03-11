package gateway

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/gateway"
)

func (mod *Module) connectTo(caller *astral.Identity, target *astral.Identity, network string) (socket gateway.Socket, err error) {
	if !mod.canGateway(caller) {
		return socket, gateway.ErrUnauthorized
	}

	endpoint, err := mod.getGatewayEndpoint(mod.ctx, network)
	if err != nil {
		return gateway.Socket{}, err
	}

	binder, ok := mod.binderByIdentity(target)
	if !ok {
		return socket, gateway.ErrTargetNotReachable
	}

	reserved, ok := binder.take()
	if !ok {
		return socket, gateway.ErrTargetNotReachable
	}

	nonce := astral.NewNonce()
	client := &client{
		Identity: caller,
		Nonce:    nonce,
		Target:   target,
		pipeTo:   reserved,
	}

	mod.connecting.Add(client)

	return gateway.Socket{
		Nonce:    nonce,
		Endpoint: endpoint,
	}, nil
}
