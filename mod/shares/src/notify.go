package shares

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/sig"
	"sync"
	"time"
)

type Notification struct {
	*Module
	Identity id.Identity
	NotifyAt time.Time
	sync.Mutex
}

func (mod *Module) Notify(identity id.Identity) error {
	n, ok := mod.notify.Set(identity.String(), &Notification{
		Module:   mod,
		Identity: identity,
		NotifyAt: time.Now().Add(mod.config.NotifyDelay),
	})
	if !ok {
		n.Lock()
		n.NotifyAt = time.Now().Add(mod.config.NotifyDelay)
		n.Unlock()
		return nil
	}

	sig.At(&n.NotifyAt, n, func() {
		mod.tasks <- func(ctx context.Context) {
			n.Notify(ctx)
			mod.notify.Delete(identity.String())
		}
	})

	return nil
}

func (n *Notification) Notify(ctx context.Context) error {
	var query = net.NewQuery(n.node.Identity(), n.Identity, notifyServiceName)
	conn, err := net.Route(ctx, n.node.Router(), query)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}
