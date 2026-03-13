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

	b, ok := mod.binderByIdentity(target)
	if !ok {
		return socket, gateway.ErrTargetNotReachable
	}

	reserved, ok := b.takeConn()
	if !ok {
		return socket, gateway.ErrTargetNotReachable
	}

	c := &connector{
		Identity: caller,
		Nonce:    astral.NewNonce(),
		Target:   target,
		reserved: reserved,
	}

	mod.connectors.Add(c)

	go func() {
		<-time.After(connectTimeout)

		bc := c.takeReserved()
		if bc == nil {
			return
		}

		mod.connectors.Remove(c)
		if err := bc.Close(); err != nil {
			mod.log.Error("failed to close reserved conn: %v", err)
		}
	}()

	return gateway.Socket{
		Nonce:    c.Nonce,
		Endpoint: endpoint,
	}, nil
}
