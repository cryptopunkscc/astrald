package gateway

import (
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/gateway"
)

const connectTimeout = 30 * time.Second

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
		// fixme: return public err ErrCannotConnect
		return socket, gateway.ErrTargetNotReachable
	}

	nonce := astral.NewNonce()
	client := &client{
		Identity: caller,
		Nonce:    nonce,
		Target:   target,
		pipeTo:   reserved,
	}

	mod.clients.Add(client)

	go func() {
		select {
		case <-time.After(connectTimeout):
		}

		binderConn := client.takePipeTo()
		if binderConn == nil {
			return
		}

		mod.clients.Remove(client)
		err = binderConn.Close()
		if err != nil {
			mod.log.Error("failed to close binderConn: %v", err)
		}
	}()

	return gateway.Socket{
		Nonce:    nonce,
		Endpoint: endpoint,
	}, nil
}
