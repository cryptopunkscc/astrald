package shares

import (
	"context"
	"encoding/json"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/lib/desc"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/router"
	"io"
	"strconv"
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

type Sync struct {
	Updates []Update
	Time    time.Time
}

type Update struct {
	DataID  data.ID
	Removed bool
}

func (c *Consumer) Sync(ctx context.Context, since time.Time) (sync Sync, err error) {
	var query = net.NewQuery(c.caller, c.target,
		router.Query(
			"shares.sync",
			router.Params{
				"since": strconv.FormatInt(since.UnixNano(), 10),
			},
		),
	)

	conn, err := net.Route(ctx, c.mod.node.Router(), query)
	if err != nil {
		return
	}
	defer conn.Close()

	for {
		var op byte
		err = cslq.Decode(conn, "c", &op)
		if err != nil {
			return
		}

		switch op {
		case opDone: // done
			var timestamp int64
			err = cslq.Decode(conn, "q", &timestamp)
			sync.Time = time.Unix(0, timestamp)
			return

		case opAdd: // add
			var dataID data.ID
			err = cslq.Decode(conn, "v", &dataID)
			if err != nil {
				return
			}

			sync.Updates = append(sync.Updates, Update{
				DataID:  dataID,
				Removed: false,
			})

		case opRemove: // remove
			var dataID data.ID
			err = cslq.Decode(conn, "v", &dataID)
			if err != nil {
				return
			}

			sync.Updates = append(sync.Updates, Update{
				DataID:  dataID,
				Removed: true,
			})

		case opResync:
			return sync, ErrResyncRequired

		case opNotFound:
			return sync, ErrUnavailable

		default:
			return sync, ErrProtocolError
		}
	}
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
