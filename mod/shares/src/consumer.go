package shares

import (
	"context"
	"encoding/json"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/lib/arl"
	"github.com/cryptopunkscc/astrald/lib/desc"
	"github.com/cryptopunkscc/astrald/mod/sets/sync"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/router"
	"io"
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

func (c *Consumer) Describe(ctx context.Context, dataID data.ID, _ *desc.Opts) (data []desc.Data, err error) {
	var query = net.NewQuery(
		c.caller,
		c.target,
		router.Query(
			"shares.describe",
			router.Params{
				"id": dataID.String(),
			},
		),
	)

	conn, err := net.Route(ctx, c.mod.node.Router(), query)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	jbytes, err := io.ReadAll(conn)
	if err != nil {
		return nil, err
	}

	var j []JSONDescriptor
	err = json.Unmarshal(jbytes, &j)
	if err != nil {
		return nil, err
	}

	for _, i := range j {
		var d = c.mod.content.UnmarshalDescriptor(i.Type, i.Data)
		if d == nil {
			continue
		}

		data = append(data, d)
	}

	return data, nil
}

func (c *Consumer) Open(ctx context.Context, dataID data.ID, opts *storage.OpenOpts) (conn net.SecureConn, err error) {
	params := router.Params{
		"id": dataID.String(),
	}

	if opts.IdentityFilter != nil {
		if !opts.IdentityFilter(c.target) {
			return
		}
	}

	if opts.Offset != 0 {
		params.SetInt("offset", int(opts.Offset))
	}

	var query = net.NewQuery(c.caller, c.target, router.Query(readServiceName, params))

	return net.Route(ctx, c.mod.node.Router(), query)
}

func (c *Consumer) Notify(ctx context.Context) error {
	var query = net.NewQuery(c.caller, c.target, notifyServiceName)
	conn, err := net.Route(ctx, c.mod.node.Router(), query)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}
