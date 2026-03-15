package gateway

import (
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/gateway"
)

const connectTimeout = 30 * time.Second

func (mod *Module) connectTo(caller *astral.Identity, target *astral.Identity, network string) (gateway.Socket, error) {
	if !mod.canGateway(caller) {
		return gateway.Socket{}, gateway.ErrUnauthorized
	}

	endpoint, err := mod.getGatewayEndpoint(mod.ctx, network)
	if err != nil {
		return gateway.Socket{}, err
	}

	binder, ok := mod.binderByIdentity(target)
	if !ok {
		return gateway.Socket{}, gateway.ErrTargetNotReachable
	}

	reserved, ok := binder.takeConn()
	if !ok {
		return gateway.Socket{}, gateway.ErrTargetNotReachable
	}

	c := &connector{
		Identity: caller,
		Nonce:    astral.NewNonce(),
		Target:   target,
		reserved: reserved,
	}

	mod.connectors.Add(c)

	go func() {
		t := time.NewTimer(connectTimeout)
		defer t.Stop()
		<-t.C

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
