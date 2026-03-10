package gateway

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/gateway"
)

type Connecting struct {
	Identity *astral.Identity
	Target   *astral.Identity
	Nonce    astral.Nonce
}

func (mod *Module) connectTo(caller *astral.Identity, target *astral.Identity, network string) (socket gateway.Socket, err error) {
	if !mod.canGateway(caller) {
		return socket, gateway.ErrUnauthorized
	}

	endpoint, err := mod.getGatewayEndpoint(mod.ctx, network)
	if err != nil {
		return gateway.Socket{}, err
	}

	nonce := astral.NewNonce()

	_, ok := mod.binderByIdentity(target)
	if !ok {
		return socket, gateway.ErrTargetNotReachable
	}

	connecting := &Connecting{
		Identity: caller,
		Target:   target,
		Nonce:    nonce,
	}

	mod.connecting.Add(connecting)

	return gateway.Socket{
		Nonce:    nonce,
		Endpoint: endpoint,
	}, nil
}

func (mod *Module) connectingByNonce(nonce astral.Nonce) (*Connecting, bool) {
	for _, c := range mod.connecting.Clone() {
		if c.Nonce == nonce {
			return c, true
		}
	}
	return nil, false
}
