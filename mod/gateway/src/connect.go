package gateway

import (
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/gateway"
)

func (mod *Module) reserveRelay(caller *astral.Identity, target *astral.Identity, network string) (gateway.Socket, error) {
	if !mod.canGateway(caller) {
		return gateway.Socket{}, gateway.ErrUnauthorized
	}

	endpoint, err := mod.getGatewayEndpoint(mod.ctx, network)
	if err != nil {
		return gateway.Socket{}, err
	}

	node, ok := mod.registeredNodeByIdentity(target)
	if !ok {
		return gateway.Socket{}, gateway.ErrTargetNotReachable
	}

	reserved, ok := node.claimConn()
	if !ok {
		return gateway.Socket{}, gateway.ErrTargetNotReachable
	}

	c := &connector{
		Identity: caller,
		Nonce:    astral.NewNonce(),
		Target:   target,
		standby:  reserved,
	}

	mod.connectors.Add(c)

	go func() {
		t := time.NewTimer(connectTimeout)
		defer t.Stop()
		<-t.C

		bc := c.claimStandby()
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
