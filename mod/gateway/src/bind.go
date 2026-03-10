package gateway

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/gateway"
)

type Binder struct {
	Identity   *astral.Identity
	Visibility gateway.Visibility
	Nonce      astral.Nonce
	ConnPool   *binderConnPool
}

func (mod *Module) bind(ctx *astral.Context, identity *astral.Identity, visibility gateway.Visibility, network string) (*gateway.Socket, error) {
	if !mod.canGateway(identity) {
		return nil, gateway.ErrUnauthorized
	}

	endpoint, err := mod.getGatewayEndpoint(ctx, network)
	if err != nil {
		return nil, err
	}

	nonce := astral.NewNonce()

	binder := &Binder{
		Identity:   identity,
		Visibility: visibility,
		Nonce:      nonce,
		ConnPool:   newBinderConnPool(mod),
	}

	if old, ok := mod.binderByIdentity(identity); ok {
		mod.binders.Remove(old)
	}

	mod.binders.Add(binder)

	return &gateway.Socket{
		Nonce:    nonce,
		Endpoint: endpoint,
	}, nil
}

func (mod *Module) canGateway(identity *astral.Identity) bool {
	return mod.config.ActAsGateway
}

func (mod *Module) binderByNonce(nonce astral.Nonce) (*Binder, bool) {
	for _, b := range mod.binders.Clone() {
		if b.Nonce == nonce {
			return b, true
		}
	}
	return nil, false
}

func (mod *Module) binderByIdentity(identity *astral.Identity) (*Binder, bool) {
	for _, b := range mod.binders.Clone() {
		if b.Identity.IsEqual(identity) {
			return b, true
		}
	}
	return nil, false
}
