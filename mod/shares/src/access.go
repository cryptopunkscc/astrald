package shares

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/net"
)

func (mod *Module) Grant(identity id.Identity, dataID data.ID) error {
	err := mod.addToLocalShareIndex(identity, dataID)
	if err != nil {
		return err
	}

	// try to notify the identity, but ignore the result
	go mod.Notify(context.Background(), identity)

	return nil
}

func (mod *Module) Revoke(identity id.Identity, dataID data.ID) error {
	err := mod.removeFromLocalShareIndex(identity, dataID)
	if err != nil {
		return err
	}

	// try to notify the identity, but ignore the result
	go mod.Notify(context.Background(), identity)

	return nil
}

func (mod *Module) Verify(identity id.Identity, dataID data.ID) bool {
	found, err := mod.localShareIndexContains(identity, dataID)

	return (err == nil) && found
}

func (mod *Module) Notify(ctx context.Context, identity id.Identity) error {
	var query = net.NewQuery(mod.node.Identity(), identity, notifyServiceName)

	conn, err := net.Route(ctx, mod.node.Router(), query)
	if err != nil {
		return err
	}
	return conn.Close()
}
