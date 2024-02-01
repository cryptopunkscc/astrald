package shares

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
	"time"
)

const notifyDelay = time.Second * 5

func (mod *Module) Notify(identity id.Identity) error {
	if mod.notify.Add(identity.String()) != nil {
		return nil
	}

	go func() {
		defer mod.notify.Remove(identity.String())

		time.Sleep(notifyDelay)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()

		var query = net.NewQuery(mod.node.Identity(), identity, notifyServiceName)
		conn, err := net.Route(ctx, mod.node.Router(), query)
		if err == nil {
			conn.Close()
		}
	}()

	return nil
}
