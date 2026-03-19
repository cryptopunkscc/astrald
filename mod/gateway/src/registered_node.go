package gateway

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/gateway"
	"github.com/cryptopunkscc/astrald/sig"
)

// registeredNode represents a node registered as reachable through the gateway.
// Only one registration per withIdentity is allowed.
type registeredNode struct {
	Identity   *astral.Identity
	Nonce      astral.Nonce
	Visibility gateway.Visibility
	idleConns  sig.Set[*idleConn]
}

func (b *registeredNode) registerConn(conn exonet.Conn, l *log.Logger) *idleConn {
	bc := newIdleConn(conn, roleGateway, b.Identity, l)
	b.idleConns.Add(bc)
	go func() { <-bc.Done(); b.idleConns.Remove(bc) }()
	return bc
}

// claimConn atomically reserves an idle idleConn for a connector.
// Uses handoffOnce as the claim token: the first caller wins.
func (b *registeredNode) claimConn() (*idleConn, bool) {
	for _, c := range b.idleConns.Clone() {
		claimed := false
		c.handoffOnce.Do(func() {
			claimed = true
			close(c.handoffCh)
		})
		if claimed {
			return c, true
		}
	}
	return nil, false
}

func (b *registeredNode) Close() error {
	for _, c := range b.idleConns.Clone() {
		c.Close()
	}
	return nil
}

func (mod *Module) register(ctx *astral.Context, identity *astral.Identity, visibility gateway.Visibility, network string) (gateway.Socket, error) {
	if !mod.canGateway(identity) {
		return gateway.Socket{}, gateway.ErrGatewayDenied
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

func (mod *Module) unregister(identity *astral.Identity) error {
	node, ok := mod.registeredNodes.Delete(identity.String())
	if !ok {
		return gateway.ErrNodeNotRegistered
	}

	err := node.Close()
	if err != nil {
		mod.log.Error("failed to close registered node: %v", err)
	}

	targetID := identity.String()
	for _, c := range mod.connectors.Clone() {
		if c.Target.String() == targetID {
			mod.connectors.Remove(c)
			err = c.Close()
			if err != nil {
				mod.log.Error("failed to close connector: %v", err)
			}
		}
	}

	return nil
}
