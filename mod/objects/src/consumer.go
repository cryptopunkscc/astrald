package objects

import (
	"context"
	"encoding/json"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/lib/desc"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/router"
	"github.com/cryptopunkscc/astrald/object"
	"io"
)

type Consumer struct {
	mod    *Module
	caller id.Identity
	target id.Identity
}

func NewConsumer(mod *Module, caller id.Identity, target id.Identity) *Consumer {
	return &Consumer{mod: mod, caller: caller, target: target}
}

func (c *Consumer) Describe(ctx context.Context, objectID object.ID, _ *desc.Opts) (data []desc.Data, err error) {
	var query = net.NewQuery(
		c.caller,
		c.target,
		router.Query(
			describeServiceName,
			router.Params{
				"id": objectID.String(),
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
		var d = c.mod.UnmarshalDescriptor(i.Type, i.Data)
		if d == nil {
			continue
		}

		data = append(data, d)
	}

	return data, nil
}

func (c *Consumer) Open(ctx context.Context, objectID object.ID, opts *objects.OpenOpts) (conn net.SecureConn, err error) {
	params := router.Params{
		"id": objectID.String(),
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
