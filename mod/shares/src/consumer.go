package shares

import (
	"context"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/lib/arl"
	"github.com/cryptopunkscc/astrald/mod/sets/sync"
	"github.com/cryptopunkscc/astrald/astral"
	"time"
)

type Consumer struct {
	mod    *Module
	caller id.Identity
	target id.Identity
}

func NewConsumer(mod *Module, caller id.Identity, target id.Identity) *Consumer {
	return &Consumer{mod: mod, caller: caller, target: target}
}

func (c *Consumer) Sync(ctx context.Context, since time.Time) (diff sync.Diff, err error) {
	return sync.NewConsumer(
		arl.New(c.caller, c.target, "shares.sync"),
		c.mod.node.Router(),
	).Sync(ctx, since)
}

func (c *Consumer) Notify(ctx context.Context) error {
	var query = astral.NewQuery(c.caller, c.target, notifyServiceName)
	conn, err := astral.Route(ctx, c.mod.node.Router(), query)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}
